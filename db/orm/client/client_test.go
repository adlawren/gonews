package client

import (
	"errors"
	"gonews/db/orm/query"
	"gonews/db/orm/query/clause"
	"gonews/db/orm/test"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	model2 := test.Model{
		Bool:   true,
		String: "def",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.All(&matchingModels)
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model2, matchingModels[1])
}

func TestAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var model test.Model
	err := client.All(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.All(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.All(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var models []*test.IdMissingModel
	err := client.All(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestDeleteAll(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	model3 := test.Model{
		Bool:   true,
		String: "ghi",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	err = client.Save(&model3)
	assert.NoError(t, err)

	err = client.DeleteAll(&[]*test.Model{&model1, &model2})
	assert.NoError(t, err)

	var models []*test.Model
	err = client.All(&models)
	assert.Len(t, models, 1)
	test.AssertModelsEqual(t, &model3, models[0])
}

func TestDeleteAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var model test.Model
	err := client.DeleteAll(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.DeleteAll(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.DeleteAll(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestDeleteAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var models []*test.IdMissingModel
	err := client.DeleteAll(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestFind(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model := test.Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var matchingModel test.Model
	err = client.Find(
		&matchingModel,
		clause.New("where string = ?", "abc"))
	assert.NoError(t, err)

	test.AssertModelsEqual(t, &model, &matchingModel)
}

func TestFindReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var models []*test.Model
	err := client.Find(&models)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var i int
	err = client.Find(&i)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var intPtr *int
	err = client.Find(&intPtr)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))
}

func TestFindReturnsErrorIfIdIsMissing(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var model test.IdMissingModel
	err := client.Find(&model)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestFindReturnsErrorIfModelNotFound(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var matchingModel test.Model
	err := client.Find(
		&matchingModel,
		clause.New("where string = ?", "abc"))
	assert.True(t, errors.Is(err, query.ErrModelNotFound))
}

func TestFindAll(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	model2 := test.Model{
		Bool:   true,
		String: "def",
	}

	err := client.Save(&model1)
	assert.NoError(t, err)

	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.New("where string = ?", "abc"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 1)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
}

func TestFindAllReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var model test.Model
	err := client.FindAll(&model)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intSlice []int
	err = client.FindAll(&intSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))

	var intPtrSlice []*int
	err = client.FindAll(&intPtrSlice)
	assert.True(t, errors.Is(err, query.ErrInvalidModelsArg))
}

func TestFindAllReturnsErrorIfIdIsMissing(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var models []*test.IdMissingModel
	err := client.FindAll(&models)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestSaveReturnsErrorIfArgumentInvalid(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var models []*test.Model
	err := client.Save(&models)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var i int
	err = client.Save(&i)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))

	var intPtr *int
	err = client.Save(&intPtr)
	assert.True(t, errors.Is(err, query.ErrInvalidModelArg))
}

func TestSaveReturnsErrorIfIdIsMissing(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	var model test.IdMissingModel
	err := client.Save(&model)
	assert.True(t, errors.Is(err, query.ErrMissingIdField))
}

func TestSaveUpdatesExistingModel(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model := test.Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	model.String = "def"

	err = client.Save(&model)
	assert.NoError(t, err)

	var matchingModel test.Model
	err = client.Find(
		&matchingModel,
		clause.New("where string = ?", "def"))
	assert.NoError(t, err)

	test.AssertModelsEqual(t, &model, &matchingModel)
}

func TestSaveSetsID(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model := test.Model{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	assert.NotZero(t, model.ID)
}

func TestSaveSetsCreatedAtIfPresent(t *testing.T) {
	db := test.InitDB(t)

	client := New(db)

	model := test.ManagedFieldsModel{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var zeroTime time.Time
	assert.NotEqual(t, model.CreatedAt, zeroTime)
}

func TestSaveUpdatesUpdatedAtIfPresent(t *testing.T) {
	db := test.InitDB(t)

	client := New(db)

	model := test.ManagedFieldsModel{
		Bool:   true,
		String: "abc",
	}

	err := client.Save(&model)
	assert.NoError(t, err)

	var zeroTime time.Time
	assert.NotEqual(t, model.UpdatedAt, zeroTime)

	previousUpdatedAt := model.UpdatedAt

	err = client.Save(&model)
	assert.NoError(t, err)

	assert.True(t, model.UpdatedAt.After(previousUpdatedAt))
}

//// Clause tests

func TestGroupByClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	model3 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err = client.Save(&model3)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.GroupBy("string"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model2, matchingModels[1])
}

func TestInClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.Where("id"),
		clause.In(model1.ID, model2.ID))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model2, matchingModels[1])
}

func TestInnerJoinClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	secondaryModel1 := test.SecondaryModel{
		ModelID: model1.ID,
		String:  "abc",
	}
	err = client.Save(&secondaryModel1)
	assert.NoError(t, err)

	secondaryModel2 := test.SecondaryModel{
		ModelID: model1.ID,
		String:  "def",
	}
	err = client.Save(&secondaryModel2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.InnerJoin("secondary_models on models.id = secondary_models.model_id"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model1, matchingModels[1])
}

func TestLeftJoinClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	secondaryModel1 := test.SecondaryModel{
		ModelID: model1.ID,
		String:  "abc",
	}
	err = client.Save(&secondaryModel1)
	assert.NoError(t, err)

	secondaryModel2 := test.SecondaryModel{
		ModelID: model1.ID,
		String:  "def",
	}
	err = client.Save(&secondaryModel2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.LeftJoin("secondary_models on models.id = secondary_models.model_id"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 3)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model1, matchingModels[1])
	test.AssertModelsEqual(t, &model2, matchingModels[2])
}

func TestLimitClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.Limit(1))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 1)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
}

func TestOrderByClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "def",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.OrderBy("string desc"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 2)
	test.AssertModelsEqual(t, &model2, matchingModels[0])
	test.AssertModelsEqual(t, &model1, matchingModels[1])
}

func TestSelectClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model)
	assert.NoError(t, err)

	var matchingModel test.Model
	err = client.Find(
		&matchingModel,
		clause.Where("id"),
		clause.In(),
		clause.Select("id from models").Wrap())
	assert.NoError(t, err)

	test.AssertModelsEqual(t, &model, &matchingModel)
}

func TestUnionClause(t *testing.T) {
	db := test.InitDB(t)
	client := New(db)

	model1 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err := client.Save(&model1)
	assert.NoError(t, err)

	model2 := test.Model{
		Bool:   true,
		String: "abc",
	}
	err = client.Save(&model2)
	assert.NoError(t, err)

	var matchingModels []*test.Model
	err = client.FindAll(
		&matchingModels,
		clause.Union("all"),
		clause.Select("* from models"))
	assert.NoError(t, err)

	assert.Len(t, matchingModels, 4)
	test.AssertModelsEqual(t, &model1, matchingModels[0])
	test.AssertModelsEqual(t, &model2, matchingModels[1])
	test.AssertModelsEqual(t, &model1, matchingModels[2])
	test.AssertModelsEqual(t, &model2, matchingModels[3])
}
