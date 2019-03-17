// Copyright 2019 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package table

// content_layout.go stores layout calculated for a canvas size.

import (
	"errors"
	"image"
	"math"
)

// columnWidth is the width of a column in cells.
// This excludes any border, padding or spacing, i.e. this is the data portion
// only.
type columnWidth int

// contentLayout determines how the content gets placed onto the canvas.
type contentLayout struct {
	// lastCvsAr is the are of the last canvas the content was drawn on.
	// This is image.ZR if the content hasn't been drawn yet.
	lastCvsAr image.Rectangle

	// columnWidths are the widths of individual columns in the table.
	columnWidths []columnWidth

	// Details about HV lines that are the borders.
}

// newContentLayout calculates new layout for the content when drawn on a
// canvas represented with the provided area.
func newContentLayout(content *Content, cvsAr image.Rectangle) (*contentLayout, error) {
	return nil, errors.New("unimplemented")
}

// columnWidths given the content and the available canvas width returns the
// widths of individual columns.
// The argument cvsWidth is assumed to exclude space required for any border,
// padding or spacing.
func columnWidths(content *Content, cvsWidth int) []columnWidth {
	// This is similar to the rod-cutting problem, except instead of maximizing
	// the price, we're minimizing the number of rows that would have their
	// content trimmed.

	inputs := &cutCanvasInputs{
		content:  content,
		cvsWidth: cvsWidth,
		columns:  int(content.cols),
		best:     map[cutState]int{},
	}
	state := &cutState{
		colIdx:   0,
		remWidth: cvsWidth,
	}

	best := cutCanvas(inputs, state, nil)

	var res []columnWidth
	last := 0
	for _, cut := range best.cuts {
		res = append(res, columnWidth(cut-last))
		last = cut
	}
	res = append(res, columnWidth(cvsWidth-last))
	return res
}

// cutState uniquely identifies a state in the cutting process.
type cutState struct {
	// colIdx is the index of the column whose width is being determined in
	// this execution of cutCanvas.
	colIdx int

	// remWidth is the total remaining width of the canvas for the current
	// column and all the following columns.
	remWidth int
}

// bestCuts is the best result for a particular cutState.
// Used for memoization.
type bestCuts struct {
	// cost is the smallest achievable cost for the cut state.
	// This is the number of rows that will have to be trimmed.
	cost int
	// cuts are the cuts done so far to get to this state.
	cuts []int
}

// cutCanvasInputs are the inputs to the cutCanvas function.
// These are shared by all the functions in the call stack.
type cutCanvasInputs struct {
	// content is the table content.
	content *Content

	// cvsWidth is the width of the canvas that is available for the data.
	cvsWidth int

	// columns indicates the total number of columns in the table.
	columns int

	// best is a memoization on top of cutCanvas.
	// It maps cutState to the minimal cost for that state.
	best map[cutState]int
}

func cutCanvas(inputs *cutCanvasInputs, state *cutState, cuts []int) *bestCuts {
	minCost := math.MaxInt32
	var minCuts []int

	nextColIdx := state.colIdx + 1
	if nextColIdx > inputs.columns-1 {
		return &bestCuts{
			cost: trimmedRows(inputs.content, state.colIdx, state.remWidth),
			cuts: cuts,
		}
	}

	for colWidth := 1; colWidth < state.remWidth; colWidth++ {
		diff := inputs.cvsWidth - state.remWidth
		idxThisCut := diff + colWidth
		costThisCut := trimmedRows(inputs.content, state.colIdx, colWidth)
		nextState := &cutState{
			colIdx:   nextColIdx,
			remWidth: state.remWidth - colWidth,
		}
		nextCuts := append(cuts, idxThisCut)

		// Use the memoized result if available.
		var nextBest *bestCuts
		if nextCost, ok := inputs.best[*nextState]; !ok {
			nextBest = cutCanvas(inputs, nextState, nextCuts)
			inputs.best[*nextState] = nextBest.cost // Memoize.
		} else {
			nextBest = &bestCuts{
				cost: nextCost,
				cuts: nextCuts,
			}
		}

		if newMinCost := costThisCut + nextBest.cost; newMinCost < minCost {
			minCost = newMinCost
			minCuts = nextBest.cuts
		}
	}
	return &bestCuts{
		cost: minCost,
		cuts: minCuts,
	}
}

// trimmedRows returns the number of rows that will have data cells with
// trimmed content in column of the specified index if the assigned width of
// the column is colWidth.
func trimmedRows(content *Content, colIdx int, colWidth int) int {
	trimmed := 0
	for _, row := range content.rows {
		tgtCell := row.cells[colIdx]
		if !tgtCell.trimmed {
			// Cells that have wrapping enabled are never trimmed and so have
			// no influence on the calculated column widths.
			continue
		}
		if tgtCell.width > colWidth {
			trimmed++
		}
	}
	return trimmed
}