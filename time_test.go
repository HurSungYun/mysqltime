package mysqltime

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"


)

func TestTime_SetAndGetDuration(t *testing.T) {
	var mt Time
	duration := 5*time.Hour + 30*time.Minute + 15*time.Second
	err := mt.SetDuration(duration)
	assert.NoError(t, err)

	retrievedDuration, valid := mt.GetDuration()
	assert.True(t, valid)
	assert.Equal(t, duration, retrievedDuration)
}

func TestTime_String(t *testing.T) {
	testCases := []struct {
		input    time.Duration
		expected string
	}{
		{5*time.Hour + 30*time.Minute + 15*time.Second, "005:30:15"},
		{-5*time.Hour - 30*time.Minute - 15*time.Second, "-005:30:15"},
		{838*time.Hour + 59*time.Minute + 59*time.Second, "838:59:59"},
		{-838*time.Hour - 59*time.Minute - 59*time.Second, "-838:59:59"},
	}

	for _, tc := range testCases {
		mt := NewTime(tc.input)
		assert.Equal(t, tc.expected, mt.String())
	}
}

func TestParseMySQLTime(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		valid    bool
	}{
		{"005:30:15", 5*time.Hour + 30*time.Minute + 15*time.Second, true},
		{"-005:30:15", -5*time.Hour - 30*time.Minute - 15*time.Second, true},
		{"838:59:59", 838*time.Hour + 59*time.Minute + 59*time.Second, true},
		{"-838:59:59", -838*time.Hour - 59*time.Minute - 59*time.Second, true},
		{"invalid", 0, false},
	}

	for _, tc := range testCases {
		duration, err := parseMySQLTime(tc.input)
		if tc.valid {
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, duration)
		} else {
			assert.Error(t, err)
		}
	}
}

func TestTime_UnmarshalText(t *testing.T) {
	var mt Time

	err := mt.UnmarshalText([]byte("005:30:15"))
	assert.NoError(t, err)

	duration, valid := mt.GetDuration()
	assert.True(t, valid)
	assert.Equal(t, 5*time.Hour+30*time.Minute+15*time.Second, duration)
}

func TestTime_MarshalText(t *testing.T) {
	mt := NewTime(5*time.Hour + 30*time.Minute + 15*time.Second)

	data, err := mt.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "005:30:15", string(data))
}

func TestTime_Scan(t *testing.T) {
	var mt Time

	err := mt.Scan("005:30:15")
	assert.NoError(t, err)

	duration, valid := mt.GetDuration()
	assert.True(t, valid)
	assert.Equal(t, 5*time.Hour+30*time.Minute+15*time.Second, duration)

	// Test invalid input
	err = mt.Scan(123) // Non-string value
	assert.Error(t, err)
}

func TestTime_Value(t *testing.T) {
	mt := NewTime(5*time.Hour + 30*time.Minute + 15*time.Second)

	val, err := mt.Value()
	assert.NoError(t, err)
	assert.Equal(t, "005:30:15", val)

	// Test invalid value (NULL scenario)
	emptyTime := Time{}
	val, err = emptyTime.Value()
	assert.NoError(t, err)
	assert.Nil(t, val)
}

func TestTime_JSONMarshaling(t *testing.T) {
	type WorkHour struct {
		UserID        int64 `json:"user_id"`
		WorkHourStart Time  `json:"work_hour_start"`
		WorkHourEnd   Time  `json:"work_hour_end"`
	}

	workHour := WorkHour{
		UserID:        1,
		WorkHourStart: NewTime(9 * time.Hour),
		WorkHourEnd:   NewTime(17 * time.Hour),
	}

	// Test JSON marshaling
	data, err := json.Marshal(workHour)
	assert.NoError(t, err)
	assert.JSONEq(t, `{"user_id":1,"work_hour_start":"009:00:00","work_hour_end":"017:00:00"}`, string(data))

	// Test JSON unmarshaling
	var unmarshaled WorkHour
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, workHour.UserID, unmarshaled.UserID)
	assert.Equal(t, workHour.WorkHourStart.String(), unmarshaled.WorkHourStart.String())
	assert.Equal(t, workHour.WorkHourEnd.String(), unmarshaled.WorkHourEnd.String())
}

func TestMySQLIntegration(t *testing.T) {
	// Connect to the MySQL database
	dsn := "testuser:testpassword@tcp(127.0.0.1:3306)/testdb"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Wait for the database to be ready
	for i := 0; i < 10; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		t.Fatalf("failed to ping MySQL: %v", err)
	}

	// Query the data
	rows, err := db.Query("SELECT user_id, work_hour_start, work_hour_end FROM work_hour")
	assert.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var userID int64
		var start, end Time

		err := rows.Scan(&userID, &start, &end)
		assert.NoError(t, err)

		fmt.Printf("UserID: %d, Start: %s, End: %s\n", userID, start.String(), end.String())
	}

	assert.NoError(t, rows.Err())
}