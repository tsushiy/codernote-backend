package crawler

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/jinzhu/gorm"
	. "github.com/tsushiy/codernote-backend/db"
)

func fetchAPI(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad response status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func cmpAndUpdate(fetched interface{}, in interface{}, db *gorm.DB) error {
	reff := reflect.ValueOf(fetched)
	refi := reflect.ValueOf(in)
	if !reff.IsValid() || !refi.IsValid() || reff.Type() != refi.Type() {
		return errors.New("different types")
	}

	if err := db.Where(in).Take(in).Error; gorm.IsRecordNotFoundError(err) {
		if err := db.Create(fetched).Error; err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	switch fetched.(type) {
	case *Problem:
		f := fetched.(*Problem)
		i := in.(*Problem)
		(*f).No = (*i).No
	case *Contest:
		f := fetched.(*Contest)
		i := in.(*Contest)
		(*f).No = (*i).No
	}

	if !reflect.DeepEqual(fetched, in) {
		if err := db.Model(in).Update(fetched).Error; err != nil {
			return err
		}
	}
	return nil
}
