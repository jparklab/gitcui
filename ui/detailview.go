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

	"github.com/rivo/tview"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// CommitDetailView is a view to show details of a commit
type CommitDetailView interface {
	GetView() *tview.Table
	SetSelected(commit *object.Commit)
}


type commitDetailView struct {
	top TopLevelView

	columns []string
	view *tview.Table

	commit *object.Commit
	stats object.FileStats
}

////////////////////////////////////////////////////////////
// comitDetailView functions
////////////////////////////////////////////////////////////

// NewCommitDetailView creates an instance of CommitDetailView
func NewCommitDetailView(top TopLevelView) CommitDetailView {
	tableView :=  tview.NewTable().
		SetBorders(false).
		SetSelectable(
			false,	// rows
			false,	// columns
		)

	tableView.
		SetBorder(true).
		SetTitle("Commit stat")


	return &commitDetailView {
		top: top,
		columns: []string{ "file", "added", "removed" },
		view: tableView,
	}
}

func (cv *commitDetailView) GetView() *tview.Table {
	return cv.view
}

func (cv *commitDetailView) SetSelected(commit *object.Commit) {
	stats, _ := commit.Stats()

	// reset view
	tableView := cv.view

	tableView.Clear()
	for idx, col := range cv.columns {
		cell := TableFormatting.Header(
			tview.NewTableCell(col).SetSelectable(false))

		if idx == 0 {
			cell.SetExpansion(1)
		}
		tableView.SetCell(0, idx, cell)
	}

	for idx, fs := range stats {
		tableView.SetCell(
			idx+1, 0,
			tview.NewTableCell(fs.Name),
		)
		tableView.SetCell(
			idx+1, 1,
			tview.NewTableCell(fmt.Sprintf("%d", fs.Addition)).
				SetAlign(tview.AlignRight),
		)
		tableView.SetCell(
			idx+1, 2,
			tview.NewTableCell(fmt.Sprintf("%d", fs.Deletion)).
				SetAlign(tview.AlignRight),
		)
	}

	cv.commit = commit
	cv.stats = stats
}