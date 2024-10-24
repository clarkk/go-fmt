package csv

import (
	"os"
	"fmt"
	"sort"
	"slices"
	"bytes"
	"regexp"
	"strings"
	"strconv"
	"unicode/utf8"
	"encoding/csv"
	"path/filepath"
	"github.com/go-errors/errors"
	"golang.org/x/sys/unix"
	"github.com/clarkk/go-fmt/sanitize"
	"github.com/clarkk/go-util/cmd"
	"github.com/clarkk/go-util/futil"
)

const (
	BOM_UTF8		= "\xEF\xBB\xBF"
	BOM_UTF16LE		= "\xFF\xFE"
	BOM_UTF16BE		= "\xFE\xFF"
	
	MIME_XLS		= "application/vnd.ms-excel"
	MIME_XLXS		= "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	
	opt_ignore_header		= "ignore_header"
	opt_remove_empty_cols	= "remove_empty_cols"
	opt_col_integrity		= "col_integrity"
)

var (
	separators		= [3]rune{',', ';', '\t'}
	
	re_col_heading 	= regexp.MustCompile(`[^\pL\d]`)
)

type (
	Reader struct {
		options		map[string]bool
		
		tmp_dir		string
		
		src			[]byte
		separator	rune
		out 		Rows
		out_header	[]string
		
		log 		Log
	}
	
	Header 		[]string
	Rows		[]row
	Log 		[]string
	
	table struct {
		Header 	Header
		Rows	Rows
	}
	
	row struct {
		Line	int			`json:"line"`
		Row		[]string	`json:"row"`
	}
)

func NewReader(tmp_dir string) *Reader {
	return &Reader{
		options: map[string]bool{
			opt_ignore_header:		false,
			opt_remove_empty_cols:	false,
			opt_col_integrity:		false,
		},
		tmp_dir: tmp_dir,
	}
}

//	Parse file
func (r *Reader) File(file, mimetype string) (table, error){
	var err error
	r.src, err = os.ReadFile(file)
	if err != nil {
		return table{}, fmt.Errorf("Unable read CSV file: %w", err)
	}
	
	return r.parse(mimetype)
}

//	Parse bytes
func (r *Reader) Bytes(b []byte, mimetype string) (table, error){
	r.src = b
	return r.parse(mimetype)
}

//	Write source to file
func (r *Reader) Write_src(file string) error {
	dir := filepath.Dir(file)
	if _, err := os.Stat(dir); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("Unable to read directory stat: %w", err)
		}
		if err := os.MkdirAll(dir, futil.CHMOD_RWX_OWNER); err != nil {
			return fmt.Errorf("Unable to create directory: %w", err)
		}
	}
	if err := unix.Access(r.tmp_dir, unix.W_OK); err != nil {
		return fmt.Errorf("Directory not writeable: %w", err)
	}
	if err := os.WriteFile(file, r.src, futil.CHMOD_RW_OWNER); err != nil {
		return fmt.Errorf("Unable to write file: %w", err)
	}
	
	var (
		log		string
		opts 	[]string
	)
	for k, v := range r.options {
		if v {
			opts = append(opts, k)
		}
	}
	if len(opts) != 0 {
		log += "Options: "+strings.Join(opts, ", ")+"\r\n\r\n"
	}
	log += strings.Join(r.Log(), "\r\n")
	if err := os.WriteFile(file+".log", []byte(log), 0664); err != nil {
		return fmt.Errorf("Unable to write file: %w", err)
	}
	
	return nil
}

//	Ignore column header
func (r *Reader) Ignore_header() *Reader {
	r.options[opt_ignore_header] = true
	return r
}

//	Remove empty colums
func (r *Reader) Remove_empty_cols() *Reader {
	r.options[opt_remove_empty_cols] = true
	return r
}

//	Ensure colum integrity (same quantity of columns in each line)
func (r *Reader) Col_integrity() *Reader {
	r.options[opt_col_integrity] = true
	return r
}

func (r *Reader) Log() []string {
	return r.log
}

func (r *Reader) parse(mimetype string) (table, error){
	if mimetype == MIME_XLS || mimetype == MIME_XLXS {
		if err := r.convert_xls(); err != nil {
			return table{}, err
		}
	}
	
	if err := r.encoding(); err != nil {
		return table{}, err
	}
	
	read := csv.NewReader(bytes.NewBuffer(r.src))
	read.FieldsPerRecord	= -1
	read.Comma				= r.separator
	
	lines, err := read.ReadAll()
	if err != nil {
		r.log_append("Unable to parse CSV")
		return table{}, &Error{"Unable to parse CSV", err}
	}
	
	r.parse_lines(lines)
	
	if len(r.out) == 0 {
		r.log_append("CSV empty")
		return table{}, &Error{"CSV empty", nil}
	}
	
	cols := r.cols()
	
	if !r.options[opt_ignore_header] {
		if cols[0] < slices.Max(cols) {
			r.log_append("Too few column headers")
			return table{}, &Error{"Too few column headers", nil}
		}
	}
	
	if r.options[opt_col_integrity] {
		if err := r.col_integrity(cols); err != nil {
			return table{}, err
		}
	} else {
		r.fill_empty_cols(cols)
	}
	
	if r.options[opt_remove_empty_cols] {
		r.remove_empty_cols()
	}
	
	if !r.options[opt_ignore_header] {
		if err := r.check_col_header(); err != nil {
			return table{}, err
		}
		
		if r.options[opt_remove_empty_cols] {
			r.remove_empty_cols()
		}
	}
	
	r.log_append(fmt.Sprintf("Rows found: %d", len(r.out)))
	return table{
		r.out_header,
		r.out,
	}, nil
}

func (r *Reader) parse_lines(lines [][]string){
	for l, line := range lines {
		empty_line := true
		
		for c, col := range line {
			col = strings.TrimSpace(col)
			
			if col != "" {
				empty_line = false
			}
			
			line[c] = col
		}
		
		//	Remove empty rows
		if !empty_line {
			r.out = append(r.out, row{
				l,
				line,
			})
		}
	}
}

func (r *Reader) remove_empty_cols(){
	cols := make([]bool, len(r.out[0].Row))
	for _, row := range r.out {
		for i, value := range row.Row {
			if value != "" {
				cols[i] = true
			}
		}
	}
	
	for c := len(cols)-1; c >= 0; c-- {
		if cols[c] {
			continue
		}
		
		r.log_append(fmt.Sprintf("Remove empty column index: %d", c))
		if len(r.out_header) != 0 {
			r.out_header = append(r.out_header[:c], r.out_header[c+1:]...)
		}
		for i := range r.out {
			r.out[i].Row = append(r.out[i].Row[:c], r.out[i].Row[c+1:]...)
		}
	}
}

func (r *Reader) check_col_header() error {
	has_heading := true
	
	first_row := r.out[0].Row
	for _, value := range first_row {
		if value == "" {
			r.log_append("Column headers cannot be empty")
			return &Error{"Column headers cannot be empty", nil}
		}
		
		value = re_col_heading.ReplaceAllString(value, "")
		if _, err := strconv.Atoi(value); err == nil {
			has_heading	= false
		}
	}
	
	if !has_heading {
		r.log_append("Column headers in CSV required")
		return &Error{"Column headers in CSV required", nil}
	} else {
		r.log_append("Column headers found")
		r.out_header	= first_row
		r.out			= r.out[1:]
	}
	return nil
}

func (r *Reader) fill_empty_cols(cols []int){
	max := slices.Max(cols)
	for t, row := range r.out {
		l := len(row.Row)
		if max != l {
			for i := 0; i < max - l; i++ {
				r.out[t].Row = append(r.out[t].Row, "")
			}
		}
	}
}

func (r *Reader) col_integrity(cols []int) error {
	if slices.Max(cols) == slices.Min(cols) {
		return nil
	}
	r.log_append("Columns in CSV not equal")
	return &Error{"Columns in CSV not equal", nil}
}

func (r *Reader) encoding() error {
	//	Detect and strip UTF8 BOM
	if bytes.HasPrefix(r.src, []byte(BOM_UTF8)) {
		s := string(r.src[len(BOM_UTF8):])
		s = sanitize.Filter_utf8mb3(s)
		s = sanitize.Trim(s, true)
		r.src_encoding_update(s)
		r.log_append("UTF8 BOM found")
		return nil
	}
	
	if utf8.Valid(r.src) {
		s := string(r.src)
		s = sanitize.Filter_utf8mb3(s)
		s = sanitize.Trim(s, true)
		r.src_encoding_update(s)
		r.log_append("UTF8 validated")
		return nil
	}
	
	s := sanitize.Normalize_non_utf8(string(r.src))
	
	//	Encode UTF8
	out := make([]byte, len(s) * utf8.UTFMax)
	n := 0
	for _, r := range []byte(s) {
		n += utf8.EncodeRune(out[n:], rune(r))
	}
	s = string(out[:n])
	s = sanitize.Filter_utf8mb3(s)
	s = sanitize.Trim(s, true)
	r.log_append("UTF8 encoded")
	
	if non_printable := sanitize.Non_printable(s); non_printable != "" {
		len_non_printable	:= len(non_printable)
		len_total			:= len(s)
		percent				:= float32(len_non_printable) / float32(len_total) * 100
		r.log_append(fmt.Sprintf("Non-printable chars found (%d / %d = %.2f%%): %s", len_non_printable, len_total, percent, non_printable))
		return &Error{"Invalid file encoding", nil}
	}
	
	r.src_encoding_update(s)
	return nil
}

func (r *Reader) convert_xls() error {
	if r.tmp_dir == "" {
		return fmt.Errorf("Temp directory not defined")
	}
	
	if err := unix.Access(r.tmp_dir, unix.W_OK); err != nil {
		return fmt.Errorf("Temp directory not writeable: %w", err)
	}
	
	f, err := os.CreateTemp(r.tmp_dir, "xls")
	if err != nil {
		return fmt.Errorf("Unable to create temp xls file: %w", err)
	}
	file_name := f.Name()
	defer os.Remove(file_name)
	
	_, err = f.Write(r.src)
	if err != nil {
		return fmt.Errorf("Unable to write temp xls file: %w", err)
	}
	
	file_name_csv := file_name+".csv"
	c := cmd.Command{}
	if err := c.Run("ssconvert "+file_name+" "+file_name_csv); err != nil {
		r.log_append("Unable to convert XLS to CSV")
		return &Error{"Unable to convert XLS to CSV", err}
	}
	defer os.Remove(file_name_csv)
	
	if err := r.src_tmp_file(file_name_csv); err != nil {
		return err
	}
	
	r.log_append("XLS converted to CSV")
	return nil
}

func (r *Reader) get_separator(s string){
	count := map[rune]int{}
	for _, sep := range separators {
		count[sep] = strings.Count(s, string(sep))
	}
	
	keys := make([]rune, len(count))
	i := 0
	for sep := range count {
		keys[i] = sep
		i++
	}
	
	sort.Slice(keys, func(i, j int) bool {
		return count[keys[i]] > count[keys[j]]
	})
	
	r.separator = keys[0]
}

func (r *Reader) src_tmp_file(file string) error {
	var err error
	r.src, err = os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Unable to read temp csv file: %w", err)
	}
	return nil
}

func (r *Reader) src_encoding_update(s string){
	r.src = []byte(s)
	r.get_separator(s)
}

func (r *Reader) cols() []int {
	cols := make([]int, len(r.out))
	for i, row := range r.out {
		cols[i] = len(row.Row)
	}
	return cols
}

func (r *Reader) log_append(s string){
	r.log = append(r.log, s)
}