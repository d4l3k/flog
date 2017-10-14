package main

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
	Rounds []struct {
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
		} `json:"customer"`
	} `json:"rounds,omitempty"`
	DiscountType     interface{} `json:"discount_type,omitempty"`
	RoundsAttributes []struct {
		ID                   interface{} `json:"id"`
		AffiliationTypeID    int         `json:"affiliation_type_id"`
		Guest                interface{} `json:"guest"`
		Paid                 bool        `json:"paid"`
		ReservationID        interface{} `json:"reservation_id"`
		EventTicketID        interface{} `json:"event_ticket_id"`
		State                string      `json:"state"`
		UserID               int         `json:"user_id"`
		RoundLinesAttributes []struct {
			ID                    interface{} `json:"id"`
			RoundID               interface{} `json:"round_id"`
			DiscountID            interface{} `json:"discount_id"`
			DiscountableProductID interface{} `json:"discountable_product_id"`
			ProductID             int         `json:"product_id"`
			ProductRuleID         int         `json:"product_rule_id"`
			OriginalUnitPrice     int         `json:"original_unit_price"`
			UnitPrice             int         `json:"unit_price"`
			UnitQuantity          int         `json:"unit_quantity"`
		} `json:"round_lines_attributes"`
	} `json:"rounds_attributes,omitempty"`
	AgreedOnTerms bool `json:"agreed_on_terms,omitempty"`
}
