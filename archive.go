package facturino

import (
	"encoding/json"
	"fmt"
)

// Archive is an archived invoice document (PDF + Factur-X + XML).
type Archive struct {
	ID           string        `json:"id"`
	Object       string        `json:"object"`
	InvoiceID    string        `json:"invoiceId"`
	Hash         string        `json:"hash"`
	PreviousHash string        `json:"previousHash"`
	ArchivedAt   string        `json:"archivedAt"`
	Files        *ArchiveFiles `json:"files,omitempty"`
}

// ArchiveFiles holds download URLs for archived documents.
type ArchiveFiles struct {
	PDFURL     string `json:"pdfUrl,omitempty"`
	FacturXURL string `json:"facturxUrl,omitempty"`
	XMLURL     string `json:"xmlUrl,omitempty"`
}

// ArchiveService operates on archived invoices.
type ArchiveService struct {
	client *httpClient
}

// List returns a paginated iterator over archived invoices.
func (s *ArchiveService) List(params *ListParams) *ArchiveIterator {
	iter := newIterator[*Archive](s.client, "/archives", params, decodeArchive)
	return &ArchiveIterator{iter: iter}
}

// Get retrieves an archived invoice by its invoice ID.
func (s *ArchiveService) Get(invoiceID string) (*Archive, error) {
	var a Archive
	err := s.client.get(fmt.Sprintf("/archives/%s", invoiceID), nil, &a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// ArchiveIterator iterates over archives.
type ArchiveIterator struct {
	iter *Iterator[*Archive]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *ArchiveIterator) Next() bool { return it.iter.Next() }

// Archive returns the most recently fetched archive.
func (it *ArchiveIterator) Archive() *Archive { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *ArchiveIterator) Err() error { return it.iter.Err() }

func decodeArchive(raw json.RawMessage) (*Archive, error) {
	var a Archive
	err := json.Unmarshal(raw, &a)
	return &a, err
}
