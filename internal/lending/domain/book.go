package domain

// Book is an aggregate root representing a physical copy that can be borrowed.
// Availability is its only mutable invariant-bearing state in this outline.
type Book struct {
	id        BookID
	title     string
	available bool
}

// NewBook creates a fresh, available Book with a domain-minted identity.
func NewBook(title string) *Book {
	return &Book{id: NewBookID(), title: title, available: true}
}

// ReconstituteBook rebuilds a Book from persisted state. Repositories use this;
// application code should not.
func ReconstituteBook(id BookID, title string, available bool) *Book {
	return &Book{id: id, title: title, available: available}
}

func (b *Book) ID() BookID        { return b.id }
func (b *Book) Title() string     { return b.title }
func (b *Book) IsAvailable() bool { return b.available }

// markOnLoan flips the Book to unavailable, enforcing the availability
// invariant. It is unexported: only the domain (the Borrow service) may
// transition a Book.
func (b *Book) markOnLoan() error {
	if !b.available {
		return ErrBookUnavailable
	}
	b.available = false
	return nil
}

// markReturned makes the Book available again.
func (b *Book) markReturned() { b.available = true }
