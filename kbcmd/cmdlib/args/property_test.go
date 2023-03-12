package args

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type TestEnum string

const (
	TestEnumFOO TestEnum = "FOO"
)

func (e TestEnum) IsValid() bool {
	return e == TestEnumFOO
}

type testObj struct {
	AccountID          string
	ParentID           string
	CompanyName        *string
	XKillbillAPIKey    string
	XKillbillAPISecret string
	IsDefault          bool
	IsDefaultPtr       *bool
	UniqueID           strfmt.UUID
	UniqueIDPtr        *strfmt.UUID
	StartTime          strfmt.DateTime
	EndTime            *strfmt.DateTime
	Enum               TestEnum
	EnumPtr            *TestEnum
}

var propertyList = []Property{
	{Name: "AccountID", Required: true},
	{Name: "ParentID", Required: true},
	{Name: "CompanyName"},
	{Name: "IsDefault"},
	{Name: "IsDefaultPtr"},
	{Name: "UniqueID"},
	{Name: "UniqueIDPtr"},
	{Name: "StartTime"},
	{Name: "EndTime"},
	{Name: "Enum"},
	{Name: "EnumPtr"},
}

func TestLoadProperties(t *testing.T) {

	uuid1 := strfmt.UUID("12345678-1234-1234-1234-123456789012")
	uuid2 := strfmt.UUID("23456789-2345-2345-2345-234567890123")
	inputs := []Input{
		{Key: "AccountID", Value: "123"},
		{Key: "ParentId", Value: "333"},
		{Key: "companyname", Value: "google"},
		{Key: "isdefault", Value: "true"},
		{Key: "isdefaultPtr", Value: "false"},
		{Key: "uniqueid", Value: string(uuid1)},
		{Key: "uniqueIdPtr", Value: string(uuid2)},
		{Key: "startTime", Value: "2018-01-02T00:00:00Z"},
		{Key: "endTime", Value: "2018-02-03T00:00:00Z"},
		{Key: "enum", Value: "FOO"},
		{Key: "enumptr", Value: "FOO"},
	}
	obj := testObj{}
	err := loadPropertiesFromInput(&obj, propertyList, inputs)
	if err != nil {
		t.Fatal(err)
	}

	companyName := "google"
	isDefault := false
	enumVal := TestEnumFOO
	exp := testObj{
		AccountID:    "123",
		ParentID:     "333",
		CompanyName:  &companyName,
		IsDefault:    true,
		IsDefaultPtr: &isDefault,
		UniqueID:     strfmt.UUID(uuid1),
		UniqueIDPtr:  &uuid2,
		StartTime:    obj.StartTime,
		EndTime:      obj.EndTime,
		Enum:         "FOO",
		EnumPtr:      &enumVal,
	}
	if diff := cmp.Diff(exp, obj, cmpopts.IgnoreUnexported(strfmt.DateTime{})); diff != "" {
		t.Fatal(diff)
	}
	dt1 := time.Date(2018, 1, 2, 0, 0, 0, 0, time.UTC)
	dt2 := time.Date(2018, 2, 3, 0, 0, 0, 0, time.UTC)
	if dt1.Sub(time.Time(obj.StartTime)) != 0 {
		t.Fatalf("invalid starttime. Expecting %v, got %v", dt1, obj.StartTime)
	}
	if dt2.Sub(time.Time(*obj.EndTime)) != 0 {
		t.Fatalf("invalid endtime. Expecting %v, got %v", dt2, obj.EndTime)
	}
}

func TestLoadProperties_InvalidEnum(t *testing.T) {

	inputs := []Input{
		{Key: "AccountID", Value: "123"},
		{Key: "ParentId", Value: "333"},
		{Key: "Enum", Value: "FOO1"},
	}
	obj := testObj{}
	err := loadPropertiesFromInput(&obj, propertyList, inputs)
	if err == nil {
		t.Fatal("expecting error, got nil instead")
	}

	expError := "Invalid value FOO1 for enum"
	if err.Error() != expError {
		t.Fatalf("Expecting: %v, Got: %v", expError, err.Error())
	}
}

func TestLoadProperties_InvalidEnumPtr(t *testing.T) {
	inputs := []Input{
		{Key: "AccountID", Value: "123"},
		{Key: "ParentId", Value: "333"},
		{Key: "EnumPtr", Value: "FOO1"},
	}
	obj := testObj{}
	err := loadPropertiesFromInput(&obj, propertyList, inputs)
	if err == nil {
		t.Fatal("expecting error, got nil instead")
	}

	expError := "Invalid value FOO1 for enum"
	if err.Error() != expError {
		t.Fatalf("Expecting: %v, Got: %v", expError, err.Error())
	}

}

func TestGetProperties(t *testing.T) {
	result := GetProperties(&testObj{})
	exp := []Property{
		{Name: "AccountID"},
		{Name: "ParentID"},
		{Name: "CompanyName"},
		{Name: "IsDefault"},
		{Name: "IsDefaultPtr"},
		{Name: "UniqueID"},
		{Name: "UniqueIDPtr"},
		{Name: "StartTime"},
		{Name: "EndTime"},
		{Name: "Enum"},
		{Name: "EnumPtr"},
	}

	if diff := cmp.Diff(Properties(exp), result); diff != "" {
		t.Fatal(diff)
	}
}

func TestGenerateUsageString_Simple(t *testing.T) {
	result := GenerateUsageString(&testObj{}, propertyList)
	exp := "\n         AccountID=STRING\n         ParentID=STRING\n         [CompanyName=STRING]\n         [IsDefault={True|False}]\n         [IsDefaultPtr={True|False}]\n         [UniqueID=UUID]\n         [UniqueIDPtr=UUID]\n         [StartTime=DATETIME]\n         [EndTime=DATETIME]\n         [Enum=STRING]\n         [EnumPtr=STRING]"
	if diff := cmp.Diff(exp, result); diff != "" {
		t.Fatal(diff)
	}
}
