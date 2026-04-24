package gocsv

func (f *File) GetCol(col int) ([]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	if col < 0 || col >= f.ColCount() {
		return nil, ErrColumnOutOfRange
	}
	out := make([]string, len(f.rows))
	for i, row := range f.rows {
		if col < len(row) {
			out[i] = row[col]
		}
	}
	return out, nil
}

func (f *File) GetColByName(header string) ([]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	idx, ok := f.HeaderIndex(header)
	if !ok {
		return nil, ErrHeaderNotFound
	}
	return f.GetCol(idx)
}

func (f *File) SetColValues(col int, values []any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if col < 0 {
		return ErrColumnOutOfRange
	}
	for len(f.rows) < len(values) {
		f.rows = append(f.rows, nil)
	}
	for i, v := range values {
		_, s := detectCellType(v)
		row := f.rows[i]
		for len(row) <= col {
			row = append(row, "")
		}
		row[col] = s
		f.rows[i] = row
	}
	return nil
}

func (f *File) InsertCol(at int, header string, values []any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if at < 0 {
		return ErrColumnOutOfRange
	}
	cols := f.ColCount()
	if at > cols {
		at = cols
	}

	if f.hasHeader {
		if len(f.headers) == 0 {
			f.headers = []string{}
		}
		for len(f.headers) < at {
			f.headers = append(f.headers, "")
		}
		f.headers = append(f.headers, "")
		copy(f.headers[at+1:], f.headers[at:])
		f.headers[at] = header
	}

	for i := range f.rows {
		row := f.rows[i]
		for len(row) < at {
			row = append(row, "")
		}
		var s string
		if i < len(values) {
			_, s = detectCellType(values[i])
		}
		row = append(row, "")
		copy(row[at+1:], row[at:])
		row[at] = s
		f.rows[i] = row
	}
	return nil
}

func (f *File) AppendCol(header string, values []any) error {
	return f.InsertCol(f.ColCount(), header, values)
}

func (f *File) RemoveCol(col int) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if col < 0 || col >= f.ColCount() {
		return ErrColumnOutOfRange
	}
	if f.hasHeader && col < len(f.headers) {
		f.headers = append(f.headers[:col], f.headers[col+1:]...)
	}
	for i, row := range f.rows {
		if col < len(row) {
			f.rows[i] = append(row[:col], row[col+1:]...)
		}
	}
	return nil
}

func (f *File) RemoveColByName(header string) error {
	idx, ok := f.HeaderIndex(header)
	if !ok {
		return ErrHeaderNotFound
	}
	return f.RemoveCol(idx)
}
