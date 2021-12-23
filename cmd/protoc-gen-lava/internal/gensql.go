package internal

import (
	"github.com/pubgo/lava/proto/lava"
	"google.golang.org/protobuf/compiler/protogen"
	gp "google.golang.org/protobuf/proto"
)

func genSql(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	for _, mth := range service.Methods {
		var opts = mth.Desc.Options()
		if !gp.HasExtension(opts, lava.E_Sqlx) {
			continue
		}

		//func Code_SendCodeExec(db *gorm.DB, arg *SendCodeRequest) *gorm.DB {
		//	return db.Raw("insert into", arg)
		//	return db.Exec("insert into", arg)
		//}

		if sql, ok := gp.GetExtension(opts, lava.E_Sqlx).(*lava.Sql); ok && sql.Exec != nil {
			g.P("func ", service.GoName, "_", mth.GoName, "Exec(", "db *", sqlxCall("DB"), ",  arg *", g.QualifiedGoIdent(mth.Input.GoIdent), ")(*", sqlxCall("DB"), "){")
			g.P(`return db.Exec("`, *sql.Query, `",arg)`)
			g.P("}")
		}

		if sql, ok := gp.GetExtension(opts, lava.E_Sqlx).(*lava.Sql); ok && sql.Query != nil {
			g.P("func ", service.GoName, "_", mth.GoName, "Raw(", "db *", sqlxCall("DB"), ",  arg *", g.QualifiedGoIdent(mth.Input.GoIdent), ")(*", sqlxCall("DB"), "){")
			g.P(`return db.Exec("`, *sql.Query, `",arg)`)
			g.P("}")
		}

		//if sql, ok := gp.GetExtension(opts, lava.E_Sqlx).(*lava.Sql); ok && sql.Query != nil {
		//	g.P("func ", service.GoName, "_", mth.GoName, "Query(ctx ", contextCall("Context"), ", db *", sqlxCall("DB"), ",  arg *", g.QualifiedGoIdent(mth.Input.GoIdent), ")([]", mth.Output.GoIdent, ", error){")
		//	g.P(`var rows, err = db.NamedQueryContext(ctx,"`, *sql.Exec, `",arg)`)
		//	g.P("if err!=nil{return nil, err}")
		//	g.P("")
		//	g.P("var resp []", mth.Output.GoIdent)
		//	g.P("return resp,sqlx.StructScan(rows, &resp)")
		//	g.P("}")
		//}
	}
}
