package messages

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/efenstakes/messenger/accounts"
	"github.com/gofiber/fiber/v2"

	"github.com/kamva/mgm/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func SaveMessage(msg Message) (*Message, error) {
	message := new(Message)
	message.ID = msg.ID
	message.To = msg.To
	message.From = msg.From
	message.Text = msg.Text

	//
	// jsonM, err := json.Marshal(msg)
	// if err != nil {
	// 	return err
	// }
	// if err := json.Unmarshal(jsonM, message); err != nil {
	// 	return err
	// }

	// check if receiver exists
	if exists := accounts.AccountExists(message.To); !exists {
		print("No such account")
		return nil, errors.New("No such account")
	}

	// !!TODO
	// check that the sender is not blocked

	if err := mgm.Coll(message).Create(message); err != nil {
		return nil, errors.New("Error Saving")
	}

	return message, nil
}

func Create(c *fiber.Ctx) error {
	message := new(Message)

	// get user from locals
	var account interface{} = c.Locals("account")
	if account == nil {
		fmt.Println("Account is not set ", account)
		return c.Status(400).JSON(fiber.Map{})
	}

	sessionAccount := account.(accounts.Account)
	print("Got session account ", sessionAccount.Name)

	if err := c.BodyParser(message); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}
	message.From = sessionAccount.Slug

	id := primitive.NewObjectID()
	fmt.Println("id ", id)
	message.ID = id
	msg, err := SaveMessage(*message)
	if err != nil {
		fmt.Println("Error saving message ", err)
		return c.Status(400).JSON(fiber.Map{})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{"message": msg})
}

func GetAll(c *fiber.Ctx) error {
	messageList := []Message{}

	// get user from locals
	var account interface{} = c.Locals("account")
	if account == nil {
		fmt.Println("Account is not set ", account)
		return c.Status(400).JSON(fiber.Map{})
	}

	sessionAccount := account.(accounts.Account)
	print("Got session account ", sessionAccount.Name)

	from := sessionAccount.Slug
	to := sessionAccount.Slug

	if val := c.Query("from"); val != "" {
		from = val
	}

	if val := c.Query("to"); val != "" {
		to = val
	}

	filters := bson.M{
		"$or": []bson.M{
			{"from": from},
			{"to": to},
		},
	}

	if err := mgm.Coll(&Message{}).SimpleFind(&messageList, filters); err != nil {
		return fiber.NewError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(messageList)
}
