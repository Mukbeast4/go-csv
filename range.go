package gocsv

type CellRange struct {
	file     *File
	startCol int
	startRow int
	endCol   int
	endRow   int
	err      error
}

func (f *File) Range(rangeRef string) *CellRange {
	r := &CellRange{file: f}
	sCol, sRow, eCol, eRow, err := splitCellRange(rangeRef)
	if err != nil {
		r.err = err
		return r
	}
	if sCol > eCol {
		sCol, eCol = eCol, sCol
	}
	if sRow > eRow {
		sRow, eRow = eRow, sRow
	}
	r.startCol = sCol
	r.startRow = sRow
	r.endCol = eCol
	r.endRow = eRow
	return r
}

func (r *CellRange) Err() error {
	return r.err
}

func (r *CellRange) Values() [][]string {
	if r.err != nil {
		return nil
	}
	rows := make([][]string, 0, r.endRow-r.startRow+1)
	for row := r.startRow; row <= r.endRow; row++ {
		line := make([]string, 0, r.endCol-r.startCol+1)
		for col := r.startCol; col <= r.endCol; col++ {
			name, _ := CoordinatesToCellName(col, row)
			v, _ := r.file.GetCellValue(name)
			line = append(line, v)
		}
		rows = append(rows, line)
	}
	return rows
}

func (r *CellRange) ForEach(fn func(col, row int, value string)) {
	if r.err != nil {
		return
	}
	for row := r.startRow; row <= r.endRow; row++ {
		for col := r.startCol; col <= r.endCol; col++ {
			name, _ := CoordinatesToCellName(col, row)
			v, _ := r.file.GetCellValue(name)
			fn(col, row, v)
		}
	}
}

func (r *CellRange) SetValue(value any) *CellRange {
	if r.err != nil {
		return r
	}
	for row := r.startRow; row <= r.endRow; row++ {
		for col := r.startCol; col <= r.endCol; col++ {
			name, _ := CoordinatesToCellName(col, row)
			if err := r.file.SetCellValue(name, value); err != nil {
				r.err = err
				return r
			}
		}
	}
	return r
}

func (r *CellRange) SetValues(values [][]any) *CellRange {
	if r.err != nil {
		return r
	}
	for i, line := range values {
		row := r.startRow + i
		if row > r.endRow {
			break
		}
		for j, v := range line {
			col := r.startCol + j
			if col > r.endCol {
				break
			}
			name, _ := CoordinatesToCellName(col, row)
			if err := r.file.SetCellValue(name, v); err != nil {
				r.err = err
				return r
			}
		}
	}
	return r
}

func (r *CellRange) Dimensions() (rows, cols int) {
	if r.err != nil {
		return 0, 0
	}
	return r.endRow - r.startRow + 1, r.endCol - r.startCol + 1
}
