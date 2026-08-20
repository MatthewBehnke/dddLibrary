# Lending

The single bounded context of this library service: it governs who may borrow
which physical books and under what rules. It exists to enforce the lending
policy — availability, membership, and loan limits — in one place.

## Language

**Book**:
A single physical copy that can be borrowed. It is either available or on loan.
_Avoid_: Title, catalogue entry, item

**Member**:
A library patron who may borrow books, subject to their membership status and
current loans.
_Avoid_: User, account, borrower, customer

**Loan**:
The record that a Member has borrowed a Book, when, and when it is due back. It
is active until returned.
_Avoid_: Checkout, rental, borrowing

**Membership Status**:
Whether a Member is currently entitled to borrow. Active members may borrow;
inactive members may not.
_Avoid_: Account state, subscription

**Due Date**:
The instant by which a borrowed Book must be returned. Fixed at the loan period
past the moment of borrowing.
_Avoid_: Return date, deadline, expiry

**Loan Period**:
The fixed span a Book may be held before it is due back (14 days).
_Avoid_: Term, duration

**Loan Limit**:
The maximum number of active Loans a single Member may hold at once (5).
_Avoid_: Quota, cap

**Borrow**:
The act of a Member taking a Book on loan. It succeeds only when the membership
is active, the Member is under the loan limit, and the Book is available.
_Avoid_: Check out, take out, lend
