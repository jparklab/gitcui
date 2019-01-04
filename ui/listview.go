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
	"github.com/rivo/tview"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// CommitListView is a view to list commits
type CommitListView interface {
	GetView() *tview.Table
}

type commitListView struct {
	top TopLevelView

	view *tview.Table
	commits []*object.Commit
	noMergeCommits []*object.Commit
}

////////////////////////////////////////////////////////////
// commitListView functions
////////////////////////////////////////////////////////////

// NewCommitListView creates an instance of CommitListView
func NewCommitListView(top TopLevelView, commits []*object.Commit) CommitListView {
	tableColumns := []string{ "hash", "message" }

	tableView := tview.NewTable().
		SetBorders(false).
		SetSelectable(
			true,		// rows
			false,		// columns
		)
	tableView.
		SetBorder(true).
		SetTitle("Commits")

	for i, col := range tableColumns {
		tableView.SetCell(
			0, i,
			TableFormatting.Header(
				tview.NewTableCell(col).
					SetSelectable(false),
			),
		)
	}

	var noMergeCommits []*object.Commit
	for _, commit := range commits {
		if commit.NumParents() > 1 {
			// ignore merges
			continue
		}

		noMergeCommits = append(noMergeCommits, commit)
	}


	for idx, commit := range noMergeCommits {

		tableView.SetCell(
			idx+1, 0,
			tview.NewTableCell(commit.Hash.String()[:10]))
		
		tableView.SetCell(
			idx+1, 1,
			tview.NewTableCell(commit.Message))

		idx++
	}

	cv := commitListView{
		top: top,
		view: tableView,
		commits: commits,
		noMergeCommits: noMergeCommits,
	}

	tableView.SetSelectionChangedFunc(cv.selectionChanged)

	return &cv

}

func (cv *commitListView) GetView() *tview.Table {
	return cv.view
}

func (cv *commitListView) selectionChanged(row, column int) {
	if cv.top != nil {
		idx := row - 1
		if idx < 0 {
			idx = 0
		} else if idx >= len(cv.noMergeCommits) {
			idx = len(cv.noMergeCommits) - 1
		}

		commit := cv.noMergeCommits[idx]
		cv.top.NotifyCommitSelectionChange(commit)
	}
}
