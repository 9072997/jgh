package jgh

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

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
	if success != false {
		t.Error("Try() returned success despite panic")
	}
	if msg.(string) != "AAAAAH!" {
		t.Error("Failed to get value of panic")
	}
}

func TestWinHTTPRequest(t *testing.T) {
	resp, status, headers := WinHTTPRequest("GET", "https://jsonplaceholder.typicode.com/posts/1", nil, "")
	if status != 200 {
		t.Fail()
	}
	if len(resp) < 100 {
		t.Fail()
	}
	if len(headers) < 3 {
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

func TestStatus(t *testing.T) {
	status := Status("413 I'm a teapot")
	if status != 413 {
		t.Error("Failed to get status code")
	}

	status = Status("I'm a teapot")
	if status != -1 {
		t.Error("Did not correctly report lack of status code")
	}

	status = Status("I'm a teapot 413")
	if status != -1 {
		t.Fail()
	}

	status = Status("413")
	if status != 413 {
		t.Error("Could not parse bare status code")
	}

	status = Status("")
	if status != -1 {
		t.Error("Did not detect error on empty string")
	}
}

func TestRandomString(t *testing.T) {
	string1 := RandomString(5)
	fmt.Println("Random string 1:", string1)
	if len(string1) != 5 {
		t.Error("Random string was not the requested length")
	}

	string2 := RandomString(5)
	fmt.Println("Random string 2:", string2)
	if string1 == string2 {
		t.Error("2 random strings were the same")
	}
}

func TestMD5(t *testing.T) {
	if MD5("") != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Error("MD5 sum of empty string did not match expected value")
	}
	if MD5("Hello World") != "b10a8db164e0754105b7a99be72e3fe5" {
		t.Error("MD5 sum of test string did not match expected value")
	}
}

func TestInt64ToStr(t *testing.T) {
	var i int64 = -7
	if Int64ToStr(i) != "-7" {
		t.Error(`Int64ToStr(-7) did not return "-7"`)
	}
	i = 4567890123
	if Int64ToStr(i) != "4567890123" {
		t.Error("Int64ToStr failed when converting a number larger than 32 bits")
	}
}

func TestExpect(t *testing.T) {
	var user1, user2, user3 userStruct
	// user1
	json.Unmarshal([]byte(`
		{
			"id": 1,
			"name": "Leanne Graham",
			"username": "Bret",
			"email": "Sincere@april.biz",
			"address": {
				"street": "Kulas Light",
				"suite": "Apt. 556",
				"city": "Gwenborough",
				"zipcode": "92998-3874",
				"geo": {
					"lat": "-37.3159",
					"lng": "81.1496"
				}
			},
			"phone": "1-770-736-8031 x56442",
			"website": "hildegard.org",
			"company": {
				"name": "Romaguera-Crona",
				"catchPhrase": "Multi-layered client-server neural-net",
				"bs": "harness real-time e-markets"
			}
		}
	`), &user1)
	// user2, just like user 1
	json.Unmarshal([]byte(`
		{
			"id": 1,
			"name": "Leanne Graham",
			"username": "Bret",
			"email": "Sincere@april.biz",
			"address": {
				"street": "Kulas Light",
				"suite": "Apt. 556",
				"city": "Gwenborough",
				"zipcode": "92998-3874",
				"geo": {
					"lat": "-37.3159",
					"lng": "81.1496"
				}
			},
			"phone": "1-770-736-8031 x56442",
			"website": "hildegard.org",
			"company": {
				"name": "Romaguera-Crona",
				"catchPhrase": "Multi-layered client-server neural-net",
				"bs": "harness real-time e-markets"
			}
		}
	`), &user2)
	// user3, diffrent lat
	json.Unmarshal([]byte(`
		{
			"id": 1,
			"name": "Leanne Graham",
			"username": "Bret",
			"email": "Sincere@april.biz",
			"address": {
				"street": "Kulas Light",
				"suite": "Apt. 556",
				"city": "Gwenborough",
				"zipcode": "92998-3874",
				"geo": {
					"lat": "99.9999",
					"lng": "81.1496"
				}
			},
			"phone": "1-770-736-8031 x56442",
			"website": "hildegard.org",
			"company": {
				"name": "Romaguera-Crona",
				"catchPhrase": "Multi-layered client-server neural-net",
				"bs": "harness real-time e-markets"
			}
		}
	`), &user3)

	success, _ := Try(0, 1, false, "", func() bool {
		Expect(user1, user2, "user2")
		return true
	})
	if !success {
		t.Error("Expect reported user1 != user2")
	}

	success, msg := Try(0, 1, false, "", func() bool {
		Expect(user1, user3, "user3")
		return true
	})
	if success {
		t.Error("Expect reported user1 == user3")
	}
	if !strings.HasPrefix(msg.(string), "Expected user3 to be") {
		t.Error("Panic message does not start with given name")
	}
}
