package csv

import (
	"os"
	"fmt"
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
	
	opt_col_integrity			= "col_integrity"
	opt_remove_empty_cols		= "remove_empty_cols"
	opt_remove_overflow_cols	= "remove_overflow_cols"
	opt_optional_header			= "optional_header"
	opt_ignore_header			= "ignore_header"
)

var (
	separators = []rune{
		',',
		';',
		'\t',
	}
	
	re_col_heading = regexp.MustCompile(`[^\pL\d]`)
)

type (
	Reader struct {
		options			map[string]bool
		
		tmp_dir			string
		
		src				[]byte
		src_converted	[]byte
		src_encoded		[]byte
		
		separator		rune
		checked_header	bool
		out 			Rows
		out_header		[]string
		
		non_printable	string
		
		log 			Log
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
			opt_col_integrity:			false,
			opt_remove_empty_cols:		false,
			opt_remove_overflow_cols:	false,
			opt_optional_header:		false,
			opt_ignore_header:			false,
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
	
	log := strings.Join(r.Log(), "\r\n")
	if err := os.WriteFile(file+".log", []byte(log), 0664); err != nil {
		return fmt.Errorf("Unable to write file: %w", err)
	}
	
	return nil
}

//	Ensure colum integrity (same quantity of columns in each line)
func (r *Reader) Col_integrity() *Reader {
	r.options[opt_col_integrity] = true
	return r
}

//	Remove empty colums
func (r *Reader) Remove_empty_cols() *Reader {
	r.options[opt_remove_empty_cols] = true
	return r
}

//	Remove overflow colums
func (r *Reader) Remove_overflow_cols() *Reader {
	r.options[opt_remove_overflow_cols] = true
	return r
}

//	Optional column header
func (r *Reader) Optional_header() *Reader {
	r.options[opt_optional_header] = true
	return r
}

//	Ignore column header
func (r *Reader) Ignore_header() *Reader {
	r.options[opt_ignore_header] = true
	return r
}

func (r *Reader) Log() []string {
	return r.log
}

func (r *Reader) parse(mimetype string) (table, error){
	r.log_options()
	
	if r.options[opt_optional_header] && r.options[opt_ignore_header] {
		return table{}, &Error{"Options 'optional_header' and 'ignore_header' can not be used in conjunction", nil}
	}
	
	if r.options[opt_remove_overflow_cols] && r.options[opt_ignore_header] {
		return table{}, &Error{"Options 'remove_overflow_cols' and 'ignore_header' can not be used in conjunction", nil}
	}
	
	if r.options[opt_remove_overflow_cols] && r.options[opt_col_integrity] {
		return table{}, &Error{"Options 'remove_overflow_cols' and 'col_integrity' can not be used in conjunction", nil}
	}
	
	if mimetype == MIME_XLS || mimetype == MIME_XLXS {
		if err := r.convert_xls(); err != nil {
			return table{}, err
		}
	}
	
	if err := r.encoding(); err != nil {
		r.log_append(err.Error())
		return table{}, &Error{err.Error(), nil}
	}
	
	read := csv.NewReader(bytes.NewBuffer(r.src_encoded))
	read.FieldsPerRecord	= -1
	read.Comma				= r.separator
	
	lines, err := read.ReadAll()
	if err != nil {
		if r.non_printable != "" {
			r.log_non_printable()
			return table{}, &Error{"Invalid CSV file encoding", nil}
		}
		r.log_append("Unable to parse CSV: "+err.Error())
		return table{}, &Error{"Unable to parse CSV: "+err.Error(), err}
	}
	r.parse_lines(lines)
	
	if err := r.empty_rows_error(); err != nil {
		return table{}, err
	}
	
	cols		:= r.cols()
	cols_max	:= slices.Max(cols)
	
	if err := r.one_col_error(cols_max); err != nil {
		return table{}, err
	}
	
	//	Remove empty columns before check_header()
	if r.options[opt_remove_empty_cols] {
		r.remove_empty_cols()
		
		cols		= r.cols()
		cols_max	= slices.Max(cols)
	}
	
	if !r.options[opt_ignore_header] {
		if r.options[opt_remove_overflow_cols] {
			if r.check_header(false) == nil {
				if err := r.empty_rows_error(); err != nil {
					return table{}, err
				}
				
				if r.options[opt_remove_empty_cols] {
					r.remove_empty_cols()
					
					cols		= r.cols()
					cols_max	= slices.Max(cols)
				}
				
				r.remove_overflow_cols()
				
				cols		= r.cols()
				cols_max	= slices.Max(cols)
			}
		}
		
		if len(r.out_header) == 0 && cols[0] < cols_max {
			r.log_append("CSV has too few column headers")
			return table{}, &Error{"CSV has too few column headers", nil}
		}
	}
	
	if r.options[opt_col_integrity] {
		if cols_max != slices.Min(cols) {
			r.log_append("Columns in CSV not equal")
			return table{}, &Error{"Columns in CSV not equal", nil}
		}
	} else {
		r.fill_empty_cols(cols_max)
	}
	
	if !r.checked_header {
		//	Optional column header
		if r.options[opt_optional_header] {
			if r.check_header(false) == nil {
				if err := r.empty_rows_error(); err != nil {
					return table{}, err
				}
				
				if r.options[opt_remove_empty_cols] {
					r.remove_empty_cols()
					
					cols		= r.cols()
					cols_max	= slices.Max(cols)
				}
			}
		//	Require column header
		} else if !r.options[opt_ignore_header] {
			if err := r.check_header(true); err != nil {
				return table{}, err
			}
			
			if err := r.empty_rows_error(); err != nil {
				return table{}, err
			}
			
			if r.options[opt_remove_empty_cols] {
				r.remove_empty_cols()
				
				cols		= r.cols()
				cols_max	= slices.Max(cols)
			}
		}
	}
	
	if err := r.one_col_error(cols_max); err != nil {
		return table{}, err
	}
	
	if r.non_printable != "" {
		r.strip_non_printable()
	}
	
	r.log_append(fmt.Sprintf("Rows found: %d", len(r.out)))
	return table{
		r.out_header,
		r.out,
	}, nil
}

func (r *Reader) encoding() error {
	var src []byte
	if len(r.src_converted) != 0 {
		src = r.src_converted
	} else {
		src = r.src
	}
	
	//	Detect and strip UTF8 BOM
	if bytes.HasPrefix(src, []byte(BOM_UTF8)) {
		s := string(src[len(BOM_UTF8):])
		s = sanitize.Filter_utf8mb3(s)
		s = sanitize.Trim(s, true)
		r.log_append("UTF8 BOM found")
		return r.src_encoding(s)
	}
	
	s := string(src)
	
	//	Valid UTF8
	if utf8.Valid(src) {
		s = sanitize.Filter_utf8mb3(s)
		s = sanitize.Trim(s, true)
		r.log_append("UTF8 validated")
		return r.src_encoding(s)
	}
	
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
	return r.src_encoding(s)
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
	
	if err := r.src_convert_file(file_name_csv); err != nil {
		return err
	}
	
	r.log_append("XLS converted to CSV")
	return nil
}

func (r *Reader) strip_non_printable(){
	c := 0
	for i, value := range r.out_header {
		s := sanitize.Strip_non_printable(value)
		if s != value {
			r.out_header[i] = s
			c++
		}
	}
	for i := range r.out {
		for j, value := range r.out[i].Row {
			s := sanitize.Strip_non_printable(value)
			if s != value {
				r.out[i].Row[j] = s
				c++
			}
		}
	}
	r.log_append(fmt.Sprintf("Values replaced (non-printable): %d", c))
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
	cols_max	:= slices.Max(r.cols())
	cols		:= make([]bool, cols_max)
	for _, row := range r.out {
		for i, value := range row.Row {
			if value != "" {
				cols[i] = true
			}
		}
	}
	
	for c := cols_max - 1; c >= 0; c-- {
		if cols[c] {
			continue
		}
		
		r.log_append(fmt.Sprintf("Remove empty column: %d", c))
		if len(r.out_header) > c {
			r.out_header = append(r.out_header[:c], r.out_header[c+1:]...)
		}
		for i := range r.out {
			if len(r.out[i].Row) > c {
				r.out[i].Row = append(r.out[i].Row[:c], r.out[i].Row[c+1:]...)
			}
		}
	}
}

func (r *Reader) remove_overflow_cols(){
	cols_max := len(r.out_header)
	for i, row := range r.out {
		if len(row.Row) > cols_max {
			r.log_append(fmt.Sprintf("Remove overflow columns row: %d", i))
			r.out[i].Row = r.out[i].Row[:cols_max]
		}
	}
}

func (r *Reader) check_header(error_log bool) error {
	r.checked_header	= true
	has_heading			:= true
	
	first_row := r.out[0].Row
	for _, value := range first_row {
		if value == "" {
			if error_log {
				r.log_append("Column headers cannot be empty")
			}
			return &Error{"Column headers cannot be empty", nil}
		}
		
		value = re_col_heading.ReplaceAllString(value, "")
		if _, err := strconv.Atoi(value); err == nil {
			has_heading	= false
		}
	}
	
	if !has_heading {
		if error_log {
			r.log_append("Column headers in CSV required")
		}
		return &Error{"Column headers in CSV required", nil}
	} else {
		r.log_append("Column headers found")
		r.out_header	= first_row
		r.out			= r.out[1:]
	}
	return nil
}

func (r *Reader) fill_empty_cols(cols_max int){
	for t, row := range r.out {
		l := len(row.Row)
		if l != cols_max {
			r.log_append(fmt.Sprintf("Fill empty columns row: %d", t))
			for i := 0; i < cols_max - l; i++ {
				r.out[t].Row = append(r.out[t].Row, "")
			}
		}
	}
}

func (r *Reader) get_separator(s string) error {
	if r.get_separator_lines(s) {
		return nil
	}
	
	c := newCount_sep()
	for _, sep := range separators {
		c.count_sep(sep, strings.Count(s, string(sep)))
	}
	
	sep, err := c.get_sep()
	if err != nil {
		return err
	}
	
	r.separator = sep
	r.log_append("Separator: "+string(r.separator))
	return nil
}

func (r *Reader) get_separator_lines(s string) bool {
	c := newCount_sep()
	for _, line := range strings.Split(s, "\n") {
		if line == "" {
			continue
		}
		for _, sep := range separators {
			c.count_lines_sep(sep, strings.Count(line, string(sep)))
		}
	}
	
	sep, err := c.get_lines_sep()
	if err != nil {
		return false
	}
	
	r.separator = sep
	r.log_append("Separator (lines): "+string(r.separator))
	return true
}

func (r *Reader) src_convert_file(file string) error {
	var err error
	r.src_converted, err = os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Unable to read temp csv file: %w", err)
	}
	return nil
}

func (r *Reader) src_encoding(s string) error {
	r.src_encoded	= []byte(s)
	r.non_printable = sanitize.Non_printable(s)
	
	if s == "" {
		return fmt.Errorf("CSV empty")
	}
	
	return r.get_separator(s)
}

func (r *Reader) cols() []int {
	cols := make([]int, len(r.out))
	for i, row := range r.out {
		cols[i] = len(row.Row)
	}
	return cols
}

func (r *Reader) log_non_printable(){
	len_total			:= len(r.src_encoded)
	len_non_printable	:= len(r.non_printable)
	percent				:= float32(len_non_printable) / float32(len_total) * 100
	r.log_append(fmt.Sprintf("Non-printable chars found (%d / %d = %.2f%%): %s", len_non_printable, len_total, percent, r.non_printable))
}

func (r *Reader) log_options(){
	var opts []string
	for k, v := range r.options {
		if v {
			opts = append(opts, k)
		}
	}
	if len(opts) != 0 {
		r.log_append("Options: "+strings.Join(opts, ", "))
	}
}

func (r *Reader) log_append(s string){
	r.log = append(r.log, s)
}

func (r *Reader) empty_rows_error() error {
	if len(r.out) == 0 {
		r.log_append("CSV empty")
		return &Error{"CSV empty", nil}
	}
	return nil
}

func (r *Reader) one_col_error(cols_max int) error {
	if cols_max == 1 {
		r.log_append("CSV must have more than one column")
		return &Error{"CSV must have more than one column", nil}
	}
	return nil
}