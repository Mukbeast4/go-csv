package gocsv

func (f *File) Headers() []string {
	if !f.hasHeader {
		return nil
	}
	out := make([]string, len(f.headers))
	copy(out, f.headers)
	return out
}

func (f *File) SetHeaders(headers []string) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	cp := make([]string, len(headers))
	copy(cp, headers)
	f.headers = cp
	f.hasHeader = true
	return nil
}

func (f *File) HeaderIndex(name string) (int, bool) {
	for i, h := range f.headers {
		if h == name {
			return i, true
		}
	}
	return -1, false
}

func (f *File) GetByHeader(row int, header string) (string, error) {
	if err := f.checkClosed(); err != nil {
		return "", err
	}
	if !f.hasHeader {
		return "", ErrNoHeader
	}
	idx, ok := f.HeaderIndex(header)
	if !ok {
		return "", ErrHeaderNotFound
	}
	if row < 0 || row >= len(f.rows) {
		return "", ErrRowOutOfRange
	}
	r := f.rows[row]
	if idx >= len(r) {
		return "", nil
	}
	return r[idx], nil
}

func (f *File) SetByHeader(row int, header string, value any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if !f.hasHeader {
		return ErrNoHeader
	}
	idx, ok := f.HeaderIndex(header)
	if !ok {
		return ErrHeaderNotFound
	}
	for len(f.rows) <= row {
		f.rows = append(f.rows, nil)
	}
	r := f.rows[row]
	for len(r) <= idx {
		r = append(r, "")
	}
	_, s := detectCellType(value)
	r[idx] = s
	f.rows[row] = r
	return nil
}

func (f *File) AppendRecord(record map[string]any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if !f.hasHeader {
		return ErrNoHeader
	}
	row := make([]string, len(f.headers))
	for i, h := range f.headers {
		if v, ok := record[h]; ok {
			_, s := detectCellType(v)
			row[i] = s
		}
	}
	f.rows = append(f.rows, row)
	return nil
}

func (f *File) GetRecord(row int) (map[string]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	if !f.hasHeader {
		return nil, ErrNoHeader
	}
	if row < 0 || row >= len(f.rows) {
		return nil, ErrRowOutOfRange
	}
	r := f.rows[row]
	out := make(map[string]string, len(f.headers))
	for i, h := range f.headers {
		if i < len(r) {
			out[h] = r[i]
		} else {
			out[h] = ""
		}
	}
	return out, nil
}

func (f *File) GetRecords() ([]map[string]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	if !f.hasHeader {
		return nil, ErrNoHeader
	}
	out := make([]map[string]string, len(f.rows))
	for i := range f.rows {
		rec, err := f.GetRecord(i)
		if err != nil {
			return nil, err
		}
		out[i] = rec
	}
	return out, nil
}
