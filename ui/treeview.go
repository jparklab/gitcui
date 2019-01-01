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
	"sort"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/utils/merkletrie"
)

const (
	MaxOpenDepth int = 2
)

// TreeContentView is a view for contents
type TreeContentView interface {
	GetView() *tview.TreeView

	// SetSelect will update treeview with the content in the commit
	// if reference is not nil, it will be used to find changes
	SetSelected(commit *object.Commit, reference *object.Commit)
}

type treeContentView struct {
	top TopLevelView
	view *tview.TreeView
}

////////////////////////////////////////////////////////////
// tree node
////////////////////////////////////////////////////////////

// NodeColor specifies the color of a tree node
type NodeColor int8

const (
	NodeColorInserted = tcell.ColorGreen
	NodeColorDeleted = tcell.ColorRed
	NodeColorModified = tcell.ColorYellow
)

type treeNodeData struct {
	entry object.TreeEntry
	changes object.Changes
	state merkletrie.Action
}

// NewTreeNodeData creates an instance of treeNodeData
func NewTreeNodeData(entry object.TreeEntry, changes object.Changes, state merkletrie.Action) *treeNodeData {
	return &treeNodeData {
		entry: entry,
		changes: changes,
		state: state,
	}
}


////////////////////////////////////////////////////////////
// treeContentView functions
////////////////////////////////////////////////////////////

// NewTreeContentView creates an instance of TreeContentView
func NewTreeContentView(top TopLevelView, commits []*object.Commit) TreeContentView {
	treeView :=  tview.NewTreeView()
	treeView.
		SetBorder(true).
		SetTitle("Current Hash Content")

	return &treeContentView {
		top: top,
		view: treeView,
	}
}

func (tv *treeContentView) GetView() *tview.TreeView {
	return tv.view
}

// helper function to add children
func addChildren(target *tview.TreeNode, parent *object.Tree) {
	for _, e := range parent.Entries {
		node := tview.NewTreeNode(e.Name).
			SetReference(e)

		target.AddChild(node)
		if e.Mode == filemode.Dir {
			child, _ := parent.Tree(e.Name)
			addChildren(node, child)
		}
	}
}

func determineState(a, b merkletrie.Action) merkletrie.Action {
	var state merkletrie.Action = a

	switch b {
	case 0:
		if state == merkletrie.Insert || state == merkletrie.Delete {
			state = merkletrie.Modify
		}
	case merkletrie.Insert:
		if state == 0 || state == merkletrie.Insert {
			state = merkletrie.Insert
		} else {
			state = merkletrie.Modify
		}
	case merkletrie.Delete:
		if state == 0 || state == merkletrie.Delete {
			state = merkletrie.Delete
		} else {
			state = merkletrie.Modify
		}
	case merkletrie.Modify:
		state = merkletrie.Modify
	}

	return state
}

func buildTree(name string, pathComponents []string, curTree *object.Tree, refTree *object.Tree, changes object.Changes, maxOpenDepth int) *tview.TreeNode {
	node := tview.NewTreeNode(name)

	changeByPath := make(map[string] object.Changes)
	changeByFullPath := make(map[string] *object.Change)

	updateMaps := func(name string, c *object.Change) {
		changeByFullPath[name] = c

		tokens := strings.Split(name, "/")
		pathComp := tokens[len(pathComponents)]

		if l, exists := changeByPath[pathComp]; exists {
			changeByPath[pathComp] = append(l, c)
		} else {
			changeByPath[pathComp] = object.Changes{c}
		}
	}

	for _, c := range changes {
		action, _ := c.Action()
		switch action {
		case merkletrie.Insert:
			updateMaps(c.To.Name, c)
		case merkletrie.Delete:
			updateMaps(c.From.Name, c)
		case merkletrie.Modify:
			updateMaps(c.From.Name, c)
			if c.To.Name != c.From.Name {
				updateMaps(c.To.Name, c)
			}
		}
	}

	var aggState merkletrie.Action

	var subdirsMap = make(map[string] bool)
	var filesMap = make(map[string] object.TreeEntry)

	var trees []*object.Tree
	if curTree != nil {
		trees = append(trees, curTree)
	}
	if refTree != nil {
		trees = append(trees, refTree)
	}

	for _, tree := range trees {
		for _, e := range tree.Entries {
			if e.Mode == filemode.Dir {
				subdirsMap[e.Name] = true
			} else {
				if _, exists := filesMap[e.Name]; !exists {
					filesMap[e.Name] = e
				}
			}
		}
	}


	var subdirs []string
	var files []string
	for d := range subdirsMap {
		subdirs = append(subdirs, d)
	}
	for f := range filesMap {
		files = append(files, f)
	}
	sort.Strings(subdirs)
	sort.Strings(files)
	
	for _, dir := range subdirs {
		var child, refChild *object.Tree
		if curTree != nil {
			child, _ = curTree.Tree(dir)
		}
		if refTree != nil {
			refChild, _ = refTree.Tree(dir)
		}

		subChanges := changeByPath[dir]

		components := append(pathComponents, dir)
		childNode := buildTree(dir, components, child, refChild, subChanges, maxOpenDepth-1)
		node.AddChild(childNode)

		data := childNode.GetReference().(*treeNodeData)
		aggState = determineState(aggState, data.state)
	}

	for _, file := range files {
		path := strings.Join(append(pathComponents, file), "/")
		childNode := tview.NewTreeNode(file)

		var data *treeNodeData
		var state merkletrie.Action
		if c, ok := changeByFullPath[path]; ok {
			state, _ = c.Action()
			if state == merkletrie.Modify && c.From.Name != c.To.Name {
				if (c.From.Name == path) {
					state = merkletrie.Delete
					childNode.SetColor(NodeColorDeleted)
				} else {
					state = merkletrie.Insert
					childNode.SetColor(NodeColorInserted)
				}
			} else {
				switch state {
				case merkletrie.Insert:
					childNode.SetColor(NodeColorInserted)
				case merkletrie.Delete:
					childNode.SetColor(NodeColorDeleted)
				case merkletrie.Modify:
					childNode.SetColor(NodeColorModified)
				}
			}
			data = NewTreeNodeData(filesMap[file], nil, state)
		} else {
			data = NewTreeNodeData(filesMap[file], nil, 0)
		}

		childNode.SetReference(data)
		node.AddChild(childNode)
		aggState = determineState(aggState, state)
	}

	switch aggState {
	case merkletrie.Insert:
		node.SetColor(NodeColorInserted)
	case merkletrie.Delete:
		node.SetColor(NodeColorDeleted)
	case merkletrie.Modify:
		node.SetColor(NodeColorModified)
	}

	var hash plumbing.Hash
	if curTree != nil {
		hash = curTree.Hash
	}

	node.SetReference(
		NewTreeNodeData(
			object.TreeEntry{
				Name: strings.Join(pathComponents, "/"),
				Mode: filemode.Dir,
				Hash: hash,
			},
			changes, aggState),
	)

	if maxOpenDepth <= 0 && aggState == 0 {
		node.SetExpanded(false)
	}

	return node
}

// SetSelected is called when a selection is changed
func (tv *treeContentView) SetSelected(commit *object.Commit, reference *object.Commit) {
	tree, _ := commit.Tree()

	var refTree *object.Tree
	var changes object.Changes
	if reference != nil {
		refTree, _ = reference.Tree()
		object.DiffTree(tree, refTree)
		changes, _ = object.DiffTree(refTree, tree)
	}

	root := buildTree(".", []string{}, tree, refTree, changes, MaxOpenDepth)

	tv.view.SetRoot(root)
}
