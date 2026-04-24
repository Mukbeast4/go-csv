package gocsv

import (
	"time"
)

func (f *File) cellIndex(cellRef string) (int, int, error) {
	col, row, err := CellNameToCoordinates(cellRef)
	if err != nil {
		return 0, 0, err
	}
	if f.hasHeader {
		if row < 2 {
			return 0, 0, &CellError{Cell: cellRef, Err: ErrRowOutOfRange}
		}
		return col - 1, row - 2, nil
	}
	return col - 1, row - 1, nil
}

func (f *File) ensureCell(colIdx, rowIdx int) {
	for len(f.rows) <= rowIdx {
		f.rows = append(f.rows, nil)
	}
	row := f.rows[rowIdx]
	for len(row) <= colIdx {
		row = append(row, "")
	}
	f.rows[rowIdx] = row
}

func (f *File) SetCellValue(cellRef string, value any) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	col, row, err := f.cellIndex(cellRef)
	if err != nil {
		return err
	}
	_, s := detectCellType(value)
	f.ensureCell(col, row)
	f.rows[row][col] = s
	return nil
}

func (f *File) GetCellValue(cellRef string) (string, error) {
	if err := f.checkClosed(); err != nil {
		return "", err
	}
	col, row, err := f.cellIndex(cellRef)
	if err != nil {
		return "", err
	}
	if row >= len(f.rows) {
		return "", nil
	}
	r := f.rows[row]
	if col >= len(r) {
		return "", nil
	}
	return r[col], nil
}

func (f *File) SetCellStr(cellRef, value string) error {
	return f.SetCellValue(cellRef, value)
}

func (f *File) GetCellStr(cellRef string) (string, error) {
	return f.GetCellValue(cellRef)
}

func (f *File) SetCellInt(cellRef string, value int64) error {
	return f.SetCellValue(cellRef, value)
}

func (f *File) GetCellInt(cellRef string) (int64, error) {
	raw, err := f.GetCellValue(cellRef)
	if err != nil {
		return 0, err
	}
	n, err := parseInt(raw)
	if err != nil {
		return 0, &CellError{Cell: cellRef, Err: err}
	}
	return n, nil
}

func (f *File) SetCellFloat(cellRef string, value float64) error {
	return f.SetCellValue(cellRef, value)
}

func (f *File) GetCellFloat(cellRef string) (float64, error) {
	raw, err := f.GetCellValue(cellRef)
	if err != nil {
		return 0, err
	}
	n, err := parseFloat(raw)
	if err != nil {
		return 0, &CellError{Cell: cellRef, Err: err}
	}
	return n, nil
}

func (f *File) SetCellBool(cellRef string, value bool) error {
	return f.SetCellValue(cellRef, value)
}

func (f *File) GetCellBool(cellRef string) (bool, error) {
	raw, err := f.GetCellValue(cellRef)
	if err != nil {
		return false, err
	}
	b, err := parseBool(raw)
	if err != nil {
		return false, &CellError{Cell: cellRef, Err: err}
	}
	return b, nil
}

func (f *File) SetCellDate(cellRef string, value time.Time) error {
	return f.SetCellValue(cellRef, value)
}

func (f *File) GetCellDate(cellRef string) (time.Time, error) {
	raw, err := f.GetCellValue(cellRef)
	if err != nil {
		return time.Time{}, err
	}
	t, err := parseDate(raw)
	if err != nil {
		return time.Time{}, &CellError{Cell: cellRef, Err: err}
	}
	return t, nil
}

func (f *File) GetCellType(cellRef string) (CellType, error) {
	raw, err := f.GetCellValue(cellRef)
	if err != nil {
		return CellTypeEmpty, err
	}
	return inferCellType(raw), nil
}
