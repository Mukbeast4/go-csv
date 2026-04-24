package gocsv

func (f *File) GetRow(row int) ([]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	if row < 0 || row >= len(f.rows) {
		return nil, ErrRowOutOfRange
	}
	out := make([]string, len(f.rows[row]))
	copy(out, f.rows[row])
	return out, nil
}

func (f *File) GetRows() ([][]string, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	out := make([][]string, len(f.rows))
	for i, r := range f.rows {
		cp := make([]string, len(r))
		copy(cp, r)
		out[i] = cp
	}
	return out, nil
}

func (f *File) SetRowValues(row int, values []any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if row < 0 {
		return ErrRowOutOfRange
	}
	for len(f.rows) <= row {
		f.rows = append(f.rows, nil)
	}
	strs := make([]string, len(values))
	for i, v := range values {
		_, s := detectCellType(v)
		strs[i] = s
	}
	f.rows[row] = strs
	return nil
}

func (f *File) AppendRow(values []any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	strs := make([]string, len(values))
	for i, v := range values {
		_, s := detectCellType(v)
		strs[i] = s
	}
	f.rows = append(f.rows, strs)
	return nil
}

func (f *File) AppendRows(rows [][]any) error {
	for _, r := range rows {
		if err := f.AppendRow(r); err != nil {
			return err
		}
	}
	return nil
}

func (f *File) AppendStrRow(values []string) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	cp := make([]string, len(values))
	copy(cp, values)
	f.rows = append(f.rows, cp)
	return nil
}

func (f *File) InsertRow(at int, values []any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if at < 0 || at > len(f.rows) {
		return ErrRowOutOfRange
	}
	strs := make([]string, len(values))
	for i, v := range values {
		_, s := detectCellType(v)
		strs[i] = s
	}
	f.rows = append(f.rows, nil)
	copy(f.rows[at+1:], f.rows[at:])
	f.rows[at] = strs
	return nil
}

func (f *File) RemoveRow(row int) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if row < 0 || row >= len(f.rows) {
		return ErrRowOutOfRange
	}
	f.rows = append(f.rows[:row], f.rows[row+1:]...)
	return nil
}

func (f *File) ClearRows() error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	f.rows = nil
	return nil
}
