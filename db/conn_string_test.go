package db

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestConstructor(t *testing.T) {
	builder := ConnStringBuilder{
		Host:     "test.com",
		Port:     "12345",
		Username: "test_user",
		Password: "secret",
		Dbname:   "test_db",
	}

	tests := []struct {
		haveFormat string
		wantString string
	}{
		{MySQL, "test_user:secret@(test.com:12345)/test_db?charset=utf8&loc=Local&parseTime=True"},
		{PostGres, "user=test_user password=secret host=test.com port=12345 dbname=test_db"},
		{MSSQL, "sqlserver://test_user:secret@test.com:12345?database=test_db"},
	}

	for _, test := range tests {
		haveString := builder.SetFormat(test.haveFormat).Build()
		assert.Equal(t, test.wantString, haveString)
	}
}
func TestBuilder(t *testing.T) {
	haveString :=
		new(ConnStringBuilder).
			SetHost("test.com").
			SetUsername("test_user").
			SetPassword("secret").
			SetDbname("test_db").
			SetPort("12345").
			Build()
	wantString := "test_user:secret@(test.com:12345)/test_db?charset=utf8&loc=Local&parseTime=True"
	assert.Equal(t, wantString, haveString)
}
func TestSetWithConfig(t *testing.T) {
	viper := viper.New()
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath("./testdata")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	haveString := new(ConnStringBuilder).SetWithConfig(viper).Build()
	wantString := "test_user:secret@(test.com:12345)/test_db?charset=utf8&loc=Local&parseTime=True"
	assert.Equal(t, wantString, haveString)
}
