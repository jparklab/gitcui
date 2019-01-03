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
	"io"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

////////////////////////////////////////////////////////////
// global variables
////////////////////////////////////////////////////////////

// TableFormatting contains formatters
var TableFormatting struct {
	Header                func(*tview.TableCell) *tview.TableCell
	Selected              func(*tview.TableCell) *tview.TableCell
}

////////////////////////////////////////////////////////////
// types
////////////////////////////////////////////////////////////

// an implementation of TopLevelView
type topLevelView struct {
	app *tview.Application
	repo *git.Repository
	commits []*object.Commit

	diffMode DiffMode
	head *object.Commit
	curSelection *object.Commit

	listView CommitListView
	detailView CommitDetailView
	treeView TreeContentView
	diffView DiffView

	curFocusView interface{}
}

// NewTopLevelView creates an instance of TopLevelView
func NewTopLevelView(app *tview.Application, repo *git.Repository, commits []*object.Commit) TopLevelView {
	topView := topLevelView{
		app: app,
		repo: repo,
		commits: commits,
		head: commits[0],
	}

	return &topView
}

func (tv *topLevelView) NotifySelectionChange(commit *object.Commit) {
	if tv.curSelection == commit {
		return
	}

	tv.curSelection = commit

	tv.detailView.SetSelected(commit)
	tv.updateTreeView()

}

// afterViewInit is called after all children views are created
func (tv *topLevelView) afterViewInit(lv CommitListView, dv CommitDetailView, tcv TreeContentView, dfv DiffView) {
	tv.listView = lv
	tv.detailView = dv
	tv.treeView = tcv
	tv.diffView = dfv

	tv.curFocusView = lv
	tv.app.SetInputCapture(func (event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 's':
				tv.switchMode()
				tv.app.Draw()
				return nil
			}
		case tcell.KeyTab:
			tv.moveFocus(true)
			tv.app.Draw()
			return nil
		case tcell.KeyBacktab:
			tv.moveFocus(false)
			tv.app.Draw()
			return nil
		}

		return event
	})

	tv.NotifySelectionChange(tv.head)
}

func (tv *topLevelView) updateTreeView() {
	commit := tv.curSelection
	// compute diff
	var reference *object.Commit
	if tv.diffMode == DiffModeSingle {
		reference, _ = commit.Parent(0)
	} else {
		reference = tv.head
	}

	tv.treeView.SetSelected(commit, reference)

	// FIXME: for testing
	if reference != nil {
		patch, _ := reference.Patch(commit)
		filePatches := patch.FilePatches()

		if len(filePatches) > 0 {
			tv.diffView.SetFilePatch(filePatches[0])
		}
	}
}

// switchMode switches diff mode
func (tv *topLevelView) switchMode() {
	if tv.diffMode == DiffModeSingle {
		tv.diffMode = DiffModeAcc
	} else {
		tv.diffMode = DiffModeSingle
	}
	tv.updateTreeView()
}

// moveFocus moves focus to the next view
func (tv *topLevelView) moveFocus(forward bool) {
	views := []interface{} {
		tv.listView,
		tv.treeView,
		tv.diffView,
		// tv.detailView,
	}

	primitives := []tview.Primitive {
		tv.listView.GetView(),
		tv.treeView.GetView(),
		tv.diffView.GetView(),
		// tv.detailView.GetView(),
	}

	for idx, v := range views {
		if v == tv.curFocusView {
			var nextIdx int
			if (forward) {
				nextIdx = (idx + 1) % len(views)
			} else {
				nextIdx = (idx + len(views) - 1) % len(views)
			}

			tv.app.SetFocus(primitives[nextIdx])
			tv.curFocusView = views[nextIdx]
			break
		}
	}
}

////////////////////////////////////////////////////////////
// internal functions
////////////////////////////////////////////////////////////

func makeViewRoot(app *tview.Application, repo *git.Repository) {
	log.Print("Loading commit logs")
	var commits []*object.Commit
	commitIter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
	})

	if err != nil {
		log.Printf("Failed to get log: %v\n", err)
		os.Exit(1)
	}

	for idx := 1; idx < 100; idx++ {
		commit, err := commitIter.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
		}

		commits = append(commits, commit)
	}

	if len(commits) == 0 {
		log.Fatal("No commits found")
	}

	log.Print("Creating views")
	topView := NewTopLevelView(app, repo, commits)

	cv := NewCommitListView(topView, commits)
	dv := NewCommitDetailView(topView)
	tv := NewTreeContentView(topView, commits)
	dfv := NewDiffView(topView)

	topView.(*topLevelView).afterViewInit(cv, dv, tv, dfv)

	// layout views
	topPanel := tview.NewFlex().
		AddItem(cv.GetView(), 0, 1, true).
		AddItem(tv.GetView(), 0, 1, false)

	root := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(topPanel, 0, 1, true).
		AddItem(dfv.GetView(), 0, 1, false)

	const FullScreen = true
	app.SetRoot(root, FullScreen)
}

func initFormatting() {
	TableFormatting.Selected = func(t *tview.TableCell) *tview.TableCell {
		return t.
			SetTextColor(tcell.ColorWhite).
			SetBackgroundColor(tcell.ColorPurple)
	}
	TableFormatting.Header = func(t *tview.TableCell) *tview.TableCell {
		return t.
			SetAttributes(tcell.AttrBold).
			SetAlign(tview.AlignCenter)
	}
}

////////////////////////////////////////////////////////////
// public functions
////////////////////////////////////////////////////////////

// Run initializes the views, and start the event handler loop
func Run(repo *git.Repository) error {
	initFormatting()

	app := tview.NewApplication()
	makeViewRoot(app, repo)

	log.Printf("Starting application")
	if err := app.Run(); err != nil {
		panic(err)
	}

	return nil
}
