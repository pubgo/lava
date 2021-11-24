package logs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/blevesearch/bleve/v2"
	bleveHttp "github.com/blevesearch/bleve/v2/http"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/xerror"

	// import general purpose configuration
	_ "github.com/blevesearch/bleve/v2/config"
)

func searchInit(i bleve.Index, data string, addr string) {
	defer xerror.RespExit()
	log.Printf("Initializing Search on %s", data)
	registerIndexes(i, data)
	initAPI(data, addr)
}

func registerIndexes(i bleve.Index, data string) {
	bleveHttp.RegisterIndexName("bleve", i)

	//// walk the data dir and register index names
	//dirEntries, err := ioutil.ReadDir(data)
	//if err != nil {
	//	log.Fatalf("error reading data dir: %v", err)
	//}
	//
	//for _, dirInfo := range dirEntries {
	//	indexPath := data + string(os.PathSeparator) + dirInfo.Name()
	//
	//	// skip single files in data dir since a valid index is a directory that
	//	// contains multiple files
	//	if !dirInfo.IsDir() {
	//		log.Printf("not registering %s, skipping", indexPath)
	//		continue
	//	}
	//
	//	i, err := bleve.Open(indexPath)
	//	if err != nil {
	//		log.Printf("error opening index %s: %v", indexPath, err)
	//	} else {
	//		log.Printf("registered index: %s", dirInfo.Name())
	//		bleveHttp.RegisterIndexName(dirInfo.Name(), i)
	//		// set correct name in stats
	//		i.SetName(dirInfo.Name())
	//	}
	//}
}

func initAPI(data string, addr string) {
	router := chi.NewMux()

	var indexNameLookup = func(req *http.Request) string {
		fmt.Println(chi.URLParam(req, "indexName"))
		return chi.URLParam(req, "indexName")
	}
	var docIDLookup = func(req *http.Request) string {
		fmt.Println(chi.URLParam(req, "docID"))
		return chi.URLParam(req, "docID")
	}

	createIndexHandler := bleveHttp.NewCreateIndexHandler(data)
	createIndexHandler.IndexNameLookup = indexNameLookup
	router.Put("/{indexName}", createIndexHandler.ServeHTTP)

	getIndexHandler := bleveHttp.NewGetIndexHandler()
	getIndexHandler.IndexNameLookup = indexNameLookup
	router.Get("/{indexName}", getIndexHandler.ServeHTTP)

	deleteIndexHandler := bleveHttp.NewDeleteIndexHandler(data)
	deleteIndexHandler.IndexNameLookup = indexNameLookup
	router.Delete("/{indexName}", deleteIndexHandler.ServeHTTP)

	listIndexesHandler := bleveHttp.NewListIndexesHandler()
	router.Get("/", listIndexesHandler.ServeHTTP)

	docIndexHandler := bleveHttp.NewDocIndexHandler("")
	docIndexHandler.IndexNameLookup = indexNameLookup
	docIndexHandler.DocIDLookup = docIDLookup
	router.Put("/{indexName}/{docID}", docIndexHandler.ServeHTTP)

	docCountHandler := bleveHttp.NewDocCountHandler("")
	docCountHandler.IndexNameLookup = indexNameLookup
	router.Get("/{indexName}/_count", docCountHandler.ServeHTTP)

	docGetHandler := bleveHttp.NewDocGetHandler("")
	docGetHandler.IndexNameLookup = indexNameLookup
	docGetHandler.DocIDLookup = docIDLookup
	router.Get("/{indexName}/{docID}", docGetHandler.ServeHTTP)

	docDeleteHandler := bleveHttp.NewDocDeleteHandler("")
	docDeleteHandler.IndexNameLookup = indexNameLookup
	docDeleteHandler.DocIDLookup = docIDLookup
	router.Delete("/{indexName}/{docID}", docDeleteHandler.ServeHTTP)

	searchHandler := bleveHttp.NewSearchHandler("")
	searchHandler.IndexNameLookup = indexNameLookup
	router.Post("/{indexName}/_search", searchHandler.ServeHTTP)

	listFieldsHandler := bleveHttp.NewListFieldsHandler("")
	listFieldsHandler.IndexNameLookup = indexNameLookup
	router.Get("/{indexName}/_fields", listFieldsHandler.ServeHTTP)
	router.Get("/{indexName}/_fields/{fieldName}", func(writer http.ResponseWriter, req *http.Request) {
		indexName := chi.URLParam(req, "indexName")
		fieldName := chi.URLParam(req, "fieldName")
		var idx = bleveHttp.IndexByName(indexName)

		i, err := idx.Advanced()
		xerror.Panic(err)

		r, err := i.Reader()
		xerror.Panic(err)

		d, err := r.FieldDict(fieldName)
		xerror.Panic(err)

		var dd []interface{}
		de, err := d.Next()
		for err == nil && de != nil {
			fmt.Println(de)
			dd = append(dd, gin.H{"count": de.Count, "term": de.Term})
			de, err = d.Next()
		}
		xerror.Panic(err)

		writer.Header().Set("Cache-Control", "no-cache")
		writer.Header().Set("Content-type", "application/json")
		writer.WriteHeader(200)
		var ddd = json.NewEncoder(writer)
		xerror.Panic(ddd.Encode(dd))
	})

	debugHandler := bleveHttp.NewDebugDocumentHandler("")
	debugHandler.IndexNameLookup = indexNameLookup
	debugHandler.DocIDLookup = docIDLookup
	router.Get("/{indexName}/{docID}/_debug", debugHandler.ServeHTTP)

	aliasHandler := bleveHttp.NewAliasHandler()
	router.Post("/_aliases", aliasHandler.ServeHTTP)

	// start the HTTP server
	http.Handle("/", router)
	log.Printf("Listening on %v", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
