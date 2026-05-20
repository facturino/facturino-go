package facturino

import (
	"encoding/json"
	"fmt"
)

// Member is a team member in a company.
type Member struct {
	ID          string `json:"id"`
	CompanyID   string `json:"companyId"`
	UserID      string `json:"userId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	InvitedBy   string `json:"invitedBy"`
	InvitedAt   string `json:"invitedAt"`
	AcceptedAt  string `json:"acceptedAt,omitempty"`
	RevokedAt   string `json:"revokedAt,omitempty"`
	Created     string `json:"created"`
	Updated     string `json:"updated"`
}

// MemberInviteParams are the parameters for inviting a member.
type MemberInviteParams struct {
	CompanyID string `json:"-"`
	Email     string `json:"email"`
	Role      string `json:"role"`

	IdempotencyKey string `json:"-"`
}

// MemberUpdateRoleParams are the parameters for updating a member's role.
type MemberUpdateRoleParams struct {
	Role string `json:"role"`
}

// MemberService operates on team members.
type MemberService struct {
	client *httpClient
}

// List returns a paginated iterator over members for a company.
func (s *MemberService) List(companyID string, params *ListParams) *MemberIterator {
	path := fmt.Sprintf("/companies/%s/members", companyID)
	iter := newIterator[*Member](s.client, path, params, decodeMember)
	return &MemberIterator{iter: iter}
}

// Get retrieves a member by ID.
func (s *MemberService) Get(companyID, id string) (*Member, error) {
	var m Member
	err := s.client.get(fmt.Sprintf("/companies/%s/members/%s", companyID, id), nil, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Invite sends an invitation to a new member.
func (s *MemberService) Invite(params *MemberInviteParams) (*Member, error) {
	var m Member
	opts := &requestOption{}
	if params.IdempotencyKey != "" {
		opts.idempotencyKey = params.IdempotencyKey
	}
	path := fmt.Sprintf("/companies/%s/members", params.CompanyID)
	err := s.client.post(path, params, &m, opts)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// UpdateRole updates a member's role.
func (s *MemberService) UpdateRole(companyID, id string, params *MemberUpdateRoleParams) (*Member, error) {
	var m Member
	err := s.client.patch(fmt.Sprintf("/companies/%s/members/%s", companyID, id), params, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Revoke revokes a member's access.
func (s *MemberService) Revoke(companyID, id string) error {
	return s.client.del(fmt.Sprintf("/companies/%s/members/%s", companyID, id))
}

// ResendInvitation re-sends the invitation email to a pending member
// (owner/admin only).
func (s *MemberService) ResendInvitation(companyID, id string) (*Member, error) {
	var m Member
	err := s.client.post(fmt.Sprintf("/companies/%s/members/%s/resend-invitation", companyID, id), nil, &m, nil)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// MemberIterator iterates over members.
type MemberIterator struct {
	iter *Iterator[*Member]
}

// Next fetches the next item from the iterator, returning false when done.
func (it *MemberIterator) Next() bool { return it.iter.Next() }

// Member returns the most recently fetched member.
func (it *MemberIterator) Member() *Member { return it.iter.Current() }

// Err returns any error encountered during iteration.
func (it *MemberIterator) Err() error { return it.iter.Err() }

func decodeMember(raw json.RawMessage) (*Member, error) {
	var m Member
	err := json.Unmarshal(raw, &m)
	return &m, err
}
