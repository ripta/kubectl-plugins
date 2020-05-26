package utree

import (
	"sort"
	"strings"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type UnstructuredTree struct {
	ByUID      map[types.UID]unstructured.Unstructured
	ByOwnerUID map[types.UID][]types.UID
	ByRoots    map[types.UID]bool
}

// Add inserts the elements into the tree.
func (t *UnstructuredTree) Add(us ...unstructured.Unstructured) {
	t.AddSlice(us)
}

// AddSlice inserts each element of the slice into the tree.
func (t *UnstructuredTree) AddSlice(us []unstructured.Unstructured) {
	if t.ByUID == nil {
		t.ByUID = make(map[types.UID]unstructured.Unstructured)
	}
	if t.ByOwnerUID == nil {
		t.ByOwnerUID = make(map[types.UID][]types.UID)
	}
	if t.ByRoots == nil {
		t.ByRoots = make(map[types.UID]bool)
	}

	for _, u := range us {
		t.ByUID[u.GetUID()] = u
		if orefs := u.GetOwnerReferences(); len(orefs) > 0 {
			delete(t.ByRoots, u.GetUID())
			for _, r := range orefs {
				if t.ByOwnerUID[r.UID] == nil {
					t.ByOwnerUID[r.UID] = make([]types.UID, 0)
				}
				t.ByOwnerUID[r.UID] = append(t.ByOwnerUID[r.UID], u.GetUID())
			}
		} else {
			t.ByRoots[u.GetUID()] = true
		}
	}
}

type UnstructuredTreeCallback func(int, unstructured.Unstructured) error

func (t *UnstructuredTree) Roots() []unstructured.Unstructured {
	objidx := map[string]types.UID{}
	synsort := []string{}

	// Sort all root objects by their GVK then name
	for uid := range t.ByRoots {
		u := t.ByUID[uid]
		// TODO: make a custom sort
		synkey := strings.Join([]string{u.GetKind(), u.GetName()}, "\t")
		objidx[synkey] = uid
		synsort = append(synsort, synkey)
	}
	sort.Strings(synsort)

	// Get all the sorted
	roots := []unstructured.Unstructured{}
	for _, synkey := range synsort {
		uid := objidx[synkey]
		roots = append(roots, t.ByUID[uid])
	}
	return roots
}

func (t *UnstructuredTree) Walk(p unstructured.Unstructured, cb UnstructuredTreeCallback) error {
	visited := make(map[types.UID]bool)
	return t.walk(0, visited, p, cb)
}

func (t *UnstructuredTree) walk(depth int, v map[types.UID]bool, p unstructured.Unstructured, cb UnstructuredTreeCallback) error {
	uids, ok := t.ByOwnerUID[p.GetUID()]
	if !ok {
		return nil
	}

	for _, uid := range uids {
		_, ok := v[uid]
		if ok {
			return errors.Errorf("BUG: found circular ownership reference involving UID %v", uid)
		}
		v[uid] = true

		u, ok := t.ByUID[uid]
		if !ok {
			return errors.Errorf("BUG: found reference to non-existent object UID %v", uid)
		}

		if err := cb(depth, u); err != nil {
			return err
		}
		if err := t.walk(depth+1, v, u, cb); err != nil {
			return err
		}
	}

	return nil
}
