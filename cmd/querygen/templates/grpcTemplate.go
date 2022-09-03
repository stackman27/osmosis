package templates

type GrpcTemplate struct {
	ProtoPath  string
	ClientPath string
	Queries    []GrpcQuery
}

type GrpcQuery struct {
	QueryName   string
	QueryPrefix bool
}

func GrpcTemplateFromQueryYml(queryYml QueryYml) GrpcTemplate {
	GrpcQueries := []GrpcQuery{}
	for queryName, queryPrefix := range queryYml.Queries {
		GrpcQueries = append(GrpcQueries, GrpcQuery{QueryName: queryName, QueryPrefix: queryPrefix.ProtoWrapper.QueryPrefix})
	}

	return GrpcTemplate{
		ProtoPath:  queryYml.protoPath,
		ClientPath: queryYml.ClientPath,
		Queries:    GrpcQueries,
	}
}
