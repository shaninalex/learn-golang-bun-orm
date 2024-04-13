package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

// structs (models) can by composed
type TimestampColumns struct {
	CreatedAt time.Time `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedAt bun.NullTime
}

type ProductVariant struct {
	bun.BaseModel
	ID         uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name       string
	ProductID  uuid.UUID              `bun:"column:notnull,type:uuid"` // simple relation via tags
	Properties map[string]interface{} // no problems with this not basic types
	TimestampColumns
}

type Product struct {
	bun.BaseModel
	ID       uuid.UUID `bun:"id,pk,type:uuid,default:uuid_generate_v4()"`
	Name     string
	Brand    string
	Variants []*ProductVariant `bun:"rel:has-many,join:id=product_id"`
	TimestampColumns
}

// get field value from composed struct
func (p *Product) getTime() time.Time {
	return p.CreatedAt
}

func migrate(ctx context.Context, db *bun.DB) error {
	models := []interface{}{
		(*Product)(nil),
		(*ProductVariant)(nil),
	}
	if err := db.ResetModel(ctx, models...); err != nil {
		return err
	}
	return nil
}

func main() {
	ctx := context.TODO()

	dsn := "postgres://postgres:password@localhost:5432/warehouse?sslmode=disable&search_path=warehouse"
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dsn),
		// pgdriver.WithNetwork("tcp"),
		// pgdriver.WithAddr("localhost:5432"),
		// pgdriver.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		// pgdriver.WithUser("postgres"),
		// pgdriver.WithPassword("password"),
		// pgdriver.WithDatabase("warehouse"),
		// pgdriver.WithConnParams(map[string]interface{}{
		// 	"search_path": "warehouse",
		// 	"sslmode":     "disable",
		// }),
	))

	db := bun.NewDB(sqldb, pgdialect.New())
	// make uuid data type available can by done by execution query after connection
	// _, err := db.ExecContext(ctx, "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	// if err != nil {
	// 	panic(err)
	// }

	// log all sql queries
	// requires to install
	// go get github.com/uptrace/bun/extra/bundebug
	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
	))

	// migrate "tables" -> see "migrate" function in this file
	if err := migrate(ctx, db); err != nil {
		panic(err)
	}

	//
	// CREATE PRODUCT
	//
	product := &Product{
		Name:  "Sport Hat",
		Brand: "brand",
	}

	_, err := db.NewInsert().Model(product).Exec(ctx)
	if err != nil {
		panic(err)
	}
	log.Printf("Inserted product id: %s\n", product.ID)

	//
	// CREATE PRODUCT VARIANTS
	//
	variants := []ProductVariant{
		{
			Name:      "red",
			ProductID: product.ID,
			Properties: map[string]interface{}{
				"color": "red",
				"size":  "X"},
		},
		{
			Name:      "green",
			ProductID: product.ID,
			Properties: map[string]interface{}{
				"color": "green",
				"size":  "M",
			},
		},
		{
			Name:      "green",
			ProductID: product.ID,
			Properties: map[string]interface{}{
				"color": "green",
				"size":  "L",
			},
		},
	}

	_, err = db.NewInsert().Model(&variants).Exec(ctx)
	if err != nil {
		panic(err)
	}

	//
	// preload all variants in prodVariants struct
	//
	prodVariants := &Product{ID: product.ID}
	err = db.NewSelect().
		Model(prodVariants).
		Relation("Variants").
		Scan(ctx)

	if err != nil {
		panic(err)
	}
	for _, variant := range prodVariants.Variants {
		log.Println(prodVariants.Name, variant)
	}

	//
	// select only product variants
	//
	// NOTE:
	// this operation is too slow. We need to unpack every "property" and
	// look up for proper field and compare it if exists...
	prodVariants2 := []*ProductVariant{}
	err = db.NewSelect().
		Model(&prodVariants2).
		Where("name = ? AND properties->>'size' LIKE ?", "green", "L").
		Scan(ctx)

	if err != nil {
		panic(err)
	}

	log.Println("All \"green\" variants with size \"L\":")
	for _, variant := range prodVariants2 {
		log.Println(variant)
	}

	//
	// UPDATE PRODUCT
	//
	updatedProduct := &Product{
		ID:    product.ID,
		Name:  "new name",
		Brand: "new brand",
	}

	res, err := db.NewUpdate().
		Model(updatedProduct).
		OmitZero().
		Where("id = ?", product.ID).
		Exec(ctx)
	if err != nil {
		panic(err)
	}

	updRows, err := res.RowsAffected()
	if err != nil {
		panic(err)
	}

	log.Printf("rows updated: %d", updRows)
}
