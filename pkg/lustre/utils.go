package lustre

import (
	"fmt"
	"k8s.io/klog/v2"
	"strings"
	"sync"
)

const (
	separator                       = "#"
	deletes                         = "delete"
	retain                          = "retain"
	archive                         = "archive"
	volumeOperationAlreadyExistsFmt = "An operation with the given Volume ID %s already exists"
)

var supportedOnDeleteValues = []string{"", deletes, retain, archive}

func validateOnDeleteValue(onDelete string) error {
	for _, v := range supportedOnDeleteValues {
		if strings.EqualFold(v, onDelete) {
			return nil
		}
	}

	return fmt.Errorf("invalid value %s for OnDelete, supported values are %v", onDelete, supportedOnDeleteValues)
}

type InFlight struct {
	mux      *sync.Mutex
	inFlight map[string]bool
}

// NewInFlight instantiates an InFlight structure.
func NewInFlight() *InFlight {
	return &InFlight{
		mux:      &sync.Mutex{},
		inFlight: make(map[string]bool),
	}
}

// Insert inserts the entry to the current list of inflight requests.
// Returns false if the key already exists.
func (db *InFlight) Insert(key string) bool {
	db.mux.Lock()
	defer db.mux.Unlock()

	_, ok := db.inFlight[key]
	if ok {
		return false
	}

	db.inFlight[key] = true
	return true
}

// Delete removes the entry from the inFlight entries map.
// It doesn't return anything, and will do nothing if the specified key doesn't exist.
func (db *InFlight) Delete(key string) {
	db.mux.Lock()
	defer db.mux.Unlock()

	delete(db.inFlight, key)
	klog.V(4).InfoS("Volume operation finished", "key", key)
}
