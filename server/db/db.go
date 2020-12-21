package db

import (
	"context"
	"fmt"
	"log"
	"webapp/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var database = "webapp"
var commodityCollection = "commodity"
var commentCollection = "comment"
var userCollection = "user"
var cartCollection = "cart"
var TokenCollection = "token"

//DB 对数据库的操作接口
type DB interface {
	//get获取所有商品|post:上传商品
	GetAllCommodity() ([]*model.Commodity, error)
	//
	GetOneCommodity(name string) (*model.Commodity, error)
	GetCommentsForCM(com string) ([]*model.Comment, error)
	WriteComment(comment *model.Comment)
	DeleteComment(comment *model.Comment)
	UpdateComment(comment *model.Comment)
	//
	GetUsersInfo() ([]*model.User, error)
	GetAUserInfo(string) ([]*model.User, error)
	UserRegister(string, string, float64) (*model.User, error)
	GetCart(username string) (*model.Cart, error)
	WriteCart(cart *model.Cart)
	PostCommodity(commodity *model.Commodity)

	AddToken(token *model.TokenKey)
	GetAToken(user string) (*model.TokenKey, error)
}

// MongoDB is the database
type MongoDB struct {
	database *mongo.Database
}

// NewMongo init the database webapp
func NewMongo(client *mongo.Client) MongoDB {
	//tech := client.Database("tech").Collection("tech")
	webapp := client.Database(database)
	return MongoDB{database: webapp}
}

//GetAllCommodity get获取所有商品
func (m MongoDB) GetAllCommodity() ([]*model.Commodity, error) {
	//获取商品表中所有的数据
	res, err := m.database.Collection(commodityCollection).Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error while fetching commodities:", err.Error())
		return nil, err
	}

	var commodities []*model.Commodity
	err = res.All(context.TODO(), &commodities)

	fmt.Println(len(commodities))

	if err != nil {
		log.Println("Error while decoding commodities:", err.Error())
		return nil, err
	}
	return commodities, nil
}

//GetOneCommodity get one commodity
func (m MongoDB) GetOneCommodity(comName string) (*model.Commodity, error) {

	var commod model.Commodity
	err := m.database.Collection(commodityCollection).FindOne(context.Background(), bson.M{"name": comName}).Decode(&commod)
	if err != nil {
		log.Println("Errorn while fetching a commodity: ", err.Error())
		return nil, err
	}
	return &commod, nil
}

//WriteComment add a comment
func (m MongoDB) WriteComment(comment *model.Comment) {
	insertComent, err := m.database.Collection(commentCollection).InsertOne(context.Background(), *comment)
	if err != nil {
		fmt.Println("Fail to insert a comment")
	}
	println("Insert a comment of ", insertComent.InsertedID)
}

//DeleteComment delete a comment
func (m MongoDB) DeleteComment(comment *model.Comment) {
	deleComment, err := m.database.Collection(commentCollection).DeleteOne(context.Background(), bson.M{"username": comment.Username, "commodity": comment.Commodity, "comment": comment.Comment})
	if err != nil {
		log.Fatal(err)
	}
	println("Delete result: ", deleComment)
}

//UpdateComment update a comment
func (m MongoDB) UpdateComment(comment *model.Comment) {
	selector := bson.M{"username": comment.Username, "commodity": comment.Commodity}
	data := bson.M{"$set": bson.M{"comment": comment.Comment}}
	updateResult, err := m.database.Collection("comment").UpdateOne(context.Background(), selector, data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updateresult: ", updateResult)

}

//GetCommentsForCM get all comments for a commodity
func (m MongoDB) GetCommentsForCM(commodity string) ([]*model.Comment, error) {
	res, err := m.database.Collection(commentCollection).Find(context.TODO(), bson.M{"commodity": commodity})

	if err != nil {
		log.Println("Error while fetching comments:", err.Error())
		return nil, err
	}

	var comments []*model.Comment
	err = res.All(context.TODO(), &comments)

	fmt.Println(len(comments))

	if err != nil {
		log.Println("Error while decoding comments:", err.Error())
		return nil, err
	}
	return comments, nil
}

/////////////////////////////////////zjy

//GetUsersInfo get usersinfo
func (m MongoDB) GetUsersInfo() ([]*model.User, error) {
	res, err := m.database.Collection(userCollection).Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println("Error while fetching all users:", err.Error())
		return nil, err
	}

	var users []*model.User
	err = res.All(context.TODO(), &users)
	if err != nil {
		log.Println("Error while decoding all users:", err.Error())
		return nil, err
	}
	return users, nil
}

//GetAUserInfo get a userinfo
func (m MongoDB) GetAUserInfo(username string) ([]*model.User, error) {
	res, err := m.database.Collection(userCollection).Find(context.TODO(), bson.M{"username": username})
	if err != nil {
		log.Println("Error while fetching a user:", err.Error())
		return nil, err
	}

	var user []*model.User
	err = res.All(context.TODO(), &user)
	if err != nil {
		log.Println("Error while decoding a user:", err.Error())
		return nil, err
	}
	return user, nil
}

//UserRegister insert a userInfo to database
func (m MongoDB) UserRegister(un string, pw string, bl float64) (*model.User, error) {
	var user model.User

	user.Username = un
	user.Password = pw
	user.Balance = bl

	selector := bson.M{"username": un}
	updateOpts := options.Update().SetUpsert(true)
	data := bson.M{"$set": user}

	updateResult, err := m.database.Collection(cartCollection).UpdateOne(context.Background(), selector, data, updateOpts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updateresult: ", updateResult)
	return &user, nil
}

// GetCart get the cart of a user
func (m MongoDB) GetCart(username string) (*model.Cart, error) {
	var cart model.Cart
	err := m.database.Collection(cartCollection).FindOne(context.Background(), bson.M{"username": username}).Decode(&cart)
	if err != nil {
		log.Println("Errorn while fetching a cart: ", err.Error())
		return nil, err
	}
	return &cart, nil
}

//WriteCart update the cart after the user add commodity to cart or remove commodity from cart
func (m MongoDB) WriteCart(cart *model.Cart) {
	selector := bson.M{"username": cart.Username}
	updateOpts := options.Update().SetUpsert(true)

	//data := bson.M{"$set": bson.M{"comment": comment.Comment}}
	data := bson.M{"$set": cart}

	updateResult, err := m.database.Collection(cartCollection).UpdateOne(context.Background(), selector, data, updateOpts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updateresult: ", updateResult)

}

//PostCommodity update or add a commodity to the app
func (m MongoDB) PostCommodity(commodity *model.Commodity) {
	selector := bson.M{"name": commodity.Name}

	updateOpts := options.Update().SetUpsert(true)

	//data := bson.M{"$set": bson.M{"comment": comment.Comment}}
	data := bson.M{"$set": commodity}

	updateResult, err := m.database.Collection(commodityCollection).UpdateOne(context.Background(), selector, data, updateOpts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updateresult: ", updateResult)
}

//AddToken add a new token to ad
func (m MongoDB) AddToken(token *model.TokenKey) {
	selector := bson.M{"username": token.Username}
	updateOpts := options.Update().SetUpsert(true)
	data := bson.M{"$set": token}

	updateResult, err := m.database.Collection(TokenCollection).UpdateOne(context.Background(), selector, data, updateOpts)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Updateresult: ", updateResult)
}

//GetAToken  get the token key of a user
func (m MongoDB) GetAToken(user string) (*model.TokenKey, error) {
	var token model.TokenKey
	err := m.database.Collection(TokenCollection).FindOne(context.TODO(), bson.M{"username": user}).Decode(&token)
	if err != nil {
		log.Println("Error while fetching a token:", err.Error())
		return nil, err
	}
	return &token, nil
}
