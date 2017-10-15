package golfer

import (
	"fmt"

	"github.com/pkg/errors"
)

type ReservationRequest struct {
	Reservation Reservation `json:"reservation"`
}

type Reservation struct {
	ID                           int         `json:"id,omitempty"`
	ClubID                       int         `json:"club_id"`
	TeetimeID                    int         `json:"teetime_id"`
	RecurrenceID                 interface{} `json:"recurrence_id"`
	ChainID                      interface{} `json:"chain_id,omitempty"`
	State                        string      `json:"state"`
	Holes                        int         `json:"holes"`
	MadeOnline                   bool        `json:"made_online"`
	OriginReservationID          interface{} `json:"origin_reservation_id"`
	CreatedUserID                int         `json:"created_user_id"`
	CreatedAt                    string      `json:"created_at,omitempty"`
	UpdatedAt                    string      `json:"updated_at,omitempty"`
	PreCheckInChronodealChosenAt interface{} `json:"pre_check_in_chronodeal_chosen_at"`
	PreCheckedInAt               interface{} `json:"pre_checked_in_at,omitempty"`
	Source                       string      `json:"source"`
	Teetime                      struct {
		ID         int         `json:"id"`
		CourseID   int         `json:"course_id"`
		StartTime  string      `json:"start_time"`
		Date       string      `json:"date"`
		EventID    interface{} `json:"event_id"`
		Hole       int         `json:"hole"`
		Round      int         `json:"round"`
		Active     bool        `json:"active"`
		Format     string      `json:"format"`
		Blocked    bool        `json:"blocked"`
		Clone      bool        `json:"clone"`
		FreeSlots  int         `json:"free_slots"`
		CartsCount int         `json:"carts_count"`
		CreatedAt  string      `json:"created_at"`
		Departure  interface{} `json:"departure"`
	} `json:"teetime,omitempty"`
	DiscountType  interface{} `json:"discount_type,omitempty"`
	AgreedOnTerms bool        `json:"agreed_on_terms,omitempty"`

	Rounds           []Round `json:"rounds,omitempty"`
	RoundsAttributes []Round `json:"rounds_attributes,omitempty"`
}

type Round struct {
	ID                int         `json:"id"`
	AffiliationTypeID int         `json:"affiliation_type_id"`
	ClubID            int         `json:"club_id"`
	Guest             interface{} `json:"guest"`
	Paid              bool        `json:"paid"`
	ReservationID     int         `json:"reservation_id"`
	EventTicketID     interface{} `json:"event_ticket_id"`
	State             string      `json:"state"`
	UserID            int         `json:"user_id"`
	Customer          struct {
		ID        int         `json:"id"`
		ClubID    int         `json:"club_id"`
		FirstName string      `json:"first_name"`
		LastName  string      `json:"last_name"`
		Phone     string      `json:"phone"`
		Email     string      `json:"email"`
		MemberNo  string      `json:"member_no"`
		BagNumber interface{} `json:"bag_number"`
	} `json:"customer,omitempty"`

	RoundLines           []RoundLine `json:"round_lines_attributes,omitempty"`
	RoundLinesAttributes []RoundLine `json:"round_lines_attributes,omitempty"`
}

type RoundLine struct {
	ID                    interface{} `json:"id"`
	RoundID               interface{} `json:"round_id"`
	DiscountID            interface{} `json:"discount_id"`
	DiscountableProductID interface{} `json:"discountable_product_id"`
	ProductID             int         `json:"product_id"`
	ProductRuleID         int         `json:"product_rule_id"`
	PaymentTransactionID  interface{} `json:"payment_transaction_id"`
	OriginalUnitPrice     float64     `json:"original_unit_price,omitempty"`
	UnitPrice             float64     `json:"unit_price"`
	UnitQuantity          int         `json:"unit_quantity"`
	AmountSubtotal        float64     `json:"amount_subtotal,omitempty"`
	AmountTax             float64     `json:"amount_tax,omitempty"`
	AmountTotal           float64     `json:"amount_total,omitempty"`
}

func (g *Golfer) Reservations() ([]Reservation, error) {
	if err := g.ensureLoggedIn(); err != nil {
		return nil, err
	}
	url := fmt.Sprintf(reservationUpcomingAPI, g.userSession.ID, g.userSession.ID)
	var r []Reservation
	if err := g.getJSON(url, &r); err != nil {
		return nil, err
	}
	return r, nil
}

func (g *Golfer) ReservationOptions(af Affiliation, c Course, tt TeeTime, players int) (Reservation, error) {
	url := fmt.Sprintf(reservationOptionsAPI, affiliationTypeIDs(af, players), tt.ID, c.Holes)
	var opts []Reservation
	if err := g.getJSON(url, &opts); err != nil {
		return Reservation{}, err
	}
	if len(opts) == 0 {
		return Reservation{}, errors.New("no options")
	}
	return opts[0], nil
}

func (g *Golfer) Reserve(af Affiliation, c Course, tt TeeTime, players int) (Reservation, error) {
	opts, err := g.ReservationOptions(af, c, tt, players)
	if err != nil {
		return Reservation{}, err
	}

	if len(opts.Rounds) == 0 {
		return Reservation{}, errors.New("no rounds present in round options")
	}
	rla := opts.Rounds[0].RoundLines

	primary := Round{
		AffiliationTypeID:    af.AffiliationTypeID,
		State:                "reserved",
		UserID:               g.userSession.ID,
		RoundLinesAttributes: rla,
	}
	secondary := Round{
		AffiliationTypeID:    af.AffiliationTypeID,
		State:                "reserved",
		RoundLinesAttributes: rla,
	}
	res := Reservation{
		AgreedOnTerms: true,
		ClubID:        af.OrganizationID,
		Holes:         c.Holes,
		MadeOnline:    true,
		Source:        "chronogolf",
		State:         "confirmed",
		TeetimeID:     tt.ID,
		RoundsAttributes: []Round{
			primary,
		},
	}
	for i := 0; i < players-1; i++ {
		res.RoundsAttributes = append(res.RoundsAttributes, secondary)
	}

	req := ReservationRequest{
		Reservation: res,
	}

	// TODO: resp
	var resp struct{}
	if err := g.postJSON(reservationAPI, req, &resp); err != nil {
		return Reservation{}, err
	}
	return Reservation{}, nil
}
