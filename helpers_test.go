package jgh

import "testing"

type userStruct struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Address  struct {
		Street  string `json:"street"`
		Suite   string `json:"suite"`
		City    string `json:"city"`
		Zipcode string `json:"zipcode"`
		Geo     struct {
			Lat string `json:"lat"`
			Lng string `json:"lng"`
		} `json:"geo"`
	} `json:"address"`
	Phone   string `json:"phone"`
	Website string `json:"website"`
	Company struct {
		Name        string `json:"name"`
		CatchPhrase string `json:"catchPhrase"`
		Bs          string `json:"bs"`
	} `json:"company"`
}

func testInitSlice(t *testing.T) {
	var s *[]bool

	InitSlice(s, 0)
	if s == nil {
		t.Fail()
	}
	if len(*s) != 0 {
		t.Fail()
	}

	s = nil
	InitSlice(s, 5)
	if len(*s) != 5 {
		t.Fail()
	}

	InitSlice(s, 2)
	if len(*s) != 2 {
		t.Fail()
	}
}

func TestTry(t *testing.T) {
	tries := 0
	success, _ := Try(0, 7, false, "Fake Testing Function", func() bool {
		tries++
		return false
	})
	if tries != 7 {
		t.Fail()
	}
	if success != false {
		t.Fail()
	}

	tries = 0
	success, _ = Try(0, 7, false, "Fake Testing Function", func() bool {
		tries++
		return tries == 5
	})
	if tries != 5 {
		t.Fail()
	}
	if success != true {
		t.Fail()
	}

	success, msg := Try(0, 1, false, "Fake Testing Function", func() bool {
		panic("AAAAAH!")
	})
	if msg.(string) != "AAAAAH!" {
		t.Fail()
	}
	if success != false {
		t.Fail()
	}
}

func TestHTTPRequest(t *testing.T) {
	client := HTTPClient(false, false)
	resp, status := HTTPRequest(client, "GET", "https://jsonplaceholder.typicode.com/posts/1", "", "", nil, "")
	if status != 200 {
		t.Fail()
	}
	if len(resp) < 100 {
		t.Fail()
	}
}

func TestRESTRequest(t *testing.T) {
	client := HTTPClient(false, false)
	var u, uOut userStruct
	u.ID = 7
	u.Name = "foo"
	u.Username = "bar"
	u.Email = "foobar"
	u.Address.Street = "baz"
	u.Address.Suite = "foobaz"
	u.Address.City = "barbaz"
	u.Address.Zipcode = "foobarbaz"
	u.Address.Geo.Lat = "buz"
	u.Address.Geo.Lng = "foobuz"
	u.Phone = "barbuz"
	u.Website = "foobarbuz"
	u.Company.Name = "bazbuz"
	u.Company.CatchPhrase = "foobazbus"
	u.Company.Bs = "barbazbuz"
	status, reflection := RESTRequest(client, "PUT", "https://jsonplaceholder.typicode.com/users/7", "", "", nil, u, nil)
	if status != 200 {
		t.Error("Status is not 200")
	}
	if !reflection {
		t.Error("False negitive for reflection")
	}

	// this will return id:1 rather than id:7
	status, reflection = RESTRequest(client, "PUT", "https://jsonplaceholder.typicode.com/users/1", "", "", nil, u, nil)
	if status != 200 {
		t.Error("Status is not 200")
	}
	if reflection {
		t.Error("Reflection failed to detect diffrent ID")
	}

	status, reflection = RESTRequest(client, "GET", "https://jsonplaceholder.typicode.com/users/1", "", "", nil, nil, &uOut)
	if uOut.Name != "Leanne Graham" {
		t.Error("Didn't get name")
	}
	if reflection {
		t.Error("False positive for reflection")
	}
}
