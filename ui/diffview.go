/**
 *  MIT License
 *
 *  Copyright (c) 2018-2018 Ji-Young Park(jiyoung.park.dev@gmail.com)
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *      The above copyright notice and this permission notice shall be included in all
 *      copies or substantial portions of the Software.
 *
 *      THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *      IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *      FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *      AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *      LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *      OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *      SOFTWARE.
 */

package ui

import (
	"fmt"
	"strings"
	
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"gopkg.in/src-d/go-git.v4/plumbing/format/diff"
)

// DiffView is a view that shows changes for a single text file
type DiffView interface {
	GetView() *tview.Table

	SetFilePatch(patch diff.FilePatch)
}

type diffView struct {
	top TopLevelView
	view *tview.Table
}

const (
	LineColorInserted = tcell.ColorGreen
	LineColorDeleted = tcell.ColorRed
)

// ExpandTabStr is used to expand tabs into strings
const ExpandTabStr = "    "

////////////////////////////////////////////////////////////
// diffView methods
////////////////////////////////////////////////////////////

// NewDiffView creates an instance of DiffView
func NewDiffView(top TopLevelView) DiffView {
	tableView :=  tview.NewTable().
		SetSelectable(
			false,	// rows
			false,	// columns
		)

	tableView.
		SetBorder(true).
		SetTitle("File Diff")

	return &diffView {
		top: top,
		view: tableView,
	}
}

func (tv *diffView) GetView() *tview.Table {
	return tv.view
}


func (tv *diffView) SetFilePatch(patch diff.FilePatch) {
	tableView := tv.view
	tableView.Clear()

	if patch == nil {
		return
	}

	tableView.SetCell(0, 0, 
		TableFormatting.Header(
			tview.NewTableCell("line").SetSelectable(false)).
		SetExpansion(1))

	idx := 0
	for _, c := range patch.Chunks() {
		content := c.Content()
		op := c.Type()

		for _, l := range strings.Split(content, "\n") {
			// since tview does not display tabs, expand tabs to string
			l = strings.Replace(l, "\t", ExpandTabStr, -1)

			var cell *tview.TableCell
			switch op {
			case diff.Equal:
				cell = tview.NewTableCell(fmt.Sprintf(" %s", l))
			case diff.Add:
				cell = tview.NewTableCell(fmt.Sprintf("+%s", l))
				cell.SetTextColor(LineColorInserted)
			case diff.Delete:
				cell = tview.NewTableCell(fmt.Sprintf("-%s", l))
				cell.SetTextColor(LineColorDeleted)
			}

			tableView.SetCell(idx, 0, cell)
			idx++
		}
	}
	tableView.ScrollToBeginning()
}