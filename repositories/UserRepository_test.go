package repositories_test

import (
	"fmt"
	"testing"

	"carrot-market-clone-api/config"
	"carrot-market-clone-api/models"
	"carrot-market-clone-api/repositories"

	"github.com/stretchr/testify/assert"
)

func TestUserRepository(t *testing.T) {
    conf, err := config.LoadTestConfig()
    if err != nil { assert.Error(t, err) }

    db, err := conf.InitDBConnection()
    if err != nil { assert.Error(t, err) }

    r := repositories.NewUserRepositoryImpl(db)

    users := make([]models.User, 5)
    for i := 0; i< len(users); i++ {
        users[i] = models.User {
            ID: fmt.Sprintf("test id %d", i+1),
            PW: fmt.Sprintf("test pw %d", i+1),
            Email: fmt.Sprintf("test email %d", i+1),
            Phone: fmt.Sprintf("test phone %d", i+1),
            Name: fmt.Sprintf("test name %d", i+1),
            Nickname: fmt.Sprintf("test nickname %d", i+1),
            ProfileImage: fmt.Sprintf("test profileImage %d", i+1),
        }
    }

    // insert
    for i := 0; i < len(users); i++ {
        if err := r.InsertUser(&users[i]); err != nil {
            assert.Error(t, err)
        }
    }

    // update
    err = r.UpdateUser(&models.User {
        ID: users[0].ID,
        Email: "update test email",
        Phone: "update test phone",
    })
    if err != nil { assert.Error(t, err) }

    // select
    columnValueMap := map[string]string{
        "id": users[0].ID,
        "email": "update test email",
        "phone": "update test phone",
    }

    for column, value := range columnValueMap {
        exists := r.CheckUserExists(column, value); 
        assert.Equal(t, true, exists)
    }

    for column, value := range columnValueMap {
        if user, err := r.GetUser(column, value); err != nil {
            assert.Error(t, err)
        } else {
            assert.Equal(t, "update test email", user.Email)
            assert.Equal(t, "update test phone", user.Phone)
        }
    }

    // delete
    for i := 0; i < len(users); i++ {
        if err := r.DeleteUser(users[i].ID); err != nil {
            assert.Error(t, err)
        }
    }

}
