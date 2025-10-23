package add

import "gopkg.in/yaml.v3"

func generateRequestSchema(name string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "type"},
			{Kind: yaml.ScalarNode, Value: "object"},
			{Kind: yaml.ScalarNode, Value: "properties"},
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "id"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "type"},
							{Kind: yaml.ScalarNode, Value: "string"},
							{Kind: yaml.ScalarNode, Value: "example"},
							{Kind: yaml.ScalarNode, Value: "123"},
						},
					},
					{Kind: yaml.ScalarNode, Value: "name"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "type"},
							{Kind: yaml.ScalarNode, Value: "string"},
							{Kind: yaml.ScalarNode, Value: "example"},
							{Kind: yaml.ScalarNode, Value: "Example Name"},
						},
					},
				},
			},
		},
	}
}

func generateResponseSchema(name string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "type"},
			{Kind: yaml.ScalarNode, Value: "object"},
			{Kind: yaml.ScalarNode, Value: "properties"},
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "id"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "type"},
							{Kind: yaml.ScalarNode, Value: "string"},
							{Kind: yaml.ScalarNode, Value: "example"},
							{Kind: yaml.ScalarNode, Value: "123"},
						},
					},
					{Kind: yaml.ScalarNode, Value: "name"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "type"},
							{Kind: yaml.ScalarNode, Value: "string"},
							{Kind: yaml.ScalarNode, Value: "example"},
							{Kind: yaml.ScalarNode, Value: "Example Name"},
						},
					},
					{Kind: yaml.ScalarNode, Value: "success"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "type"},
							{Kind: yaml.ScalarNode, Value: "boolean"},
							{Kind: yaml.ScalarNode, Value: "example"},
							{Kind: yaml.ScalarNode, Value: "true"},
						},
					},
				},
			},
		},
	}
}

func generatePathItem(name string) *yaml.Node {
	requestRef := "#/components/schemas/" + name + "Request"
	responseRef := "#/components/schemas/" + name + "Response"
	errorRef := "#/components/schemas/Error"

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Value: "post"},
			{
				Kind: yaml.MappingNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "summary"},
					{Kind: yaml.ScalarNode, Value: name + " operation"},
					{Kind: yaml.ScalarNode, Value: "operationId"},
					{Kind: yaml.ScalarNode, Value: camelCase(name)},
					{Kind: yaml.ScalarNode, Value: "requestBody"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "required"},
							{Kind: yaml.ScalarNode, Value: "true"},
							{Kind: yaml.ScalarNode, Value: "content"},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{Kind: yaml.ScalarNode, Value: "application/json"},
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Value: "schema"},
											{
												Kind: yaml.MappingNode,
												Content: []*yaml.Node{
													{Kind: yaml.ScalarNode, Value: "$ref"},
													{Kind: yaml.ScalarNode, Value: requestRef},
												},
											},
										},
									},
								},
							},
						},
					},
					{Kind: yaml.ScalarNode, Value: "responses"},
					{
						Kind: yaml.MappingNode,
						Content: []*yaml.Node{
							{Kind: yaml.ScalarNode, Value: "200", Style: yaml.SingleQuotedStyle},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{Kind: yaml.ScalarNode, Value: "description"},
									{Kind: yaml.ScalarNode, Value: "Successful response"},
									{Kind: yaml.ScalarNode, Value: "content"},
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Value: "application/json"},
											{
												Kind: yaml.MappingNode,
												Content: []*yaml.Node{
													{Kind: yaml.ScalarNode, Value: "schema"},
													{
														Kind: yaml.MappingNode,
														Content: []*yaml.Node{
															{Kind: yaml.ScalarNode, Value: "$ref"},
															{Kind: yaml.ScalarNode, Value: responseRef},
														},
													},
												},
											},
										},
									},
								},
							},
							{Kind: yaml.ScalarNode, Value: "400", Style: yaml.SingleQuotedStyle},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{Kind: yaml.ScalarNode, Value: "description"},
									{Kind: yaml.ScalarNode, Value: "Bad request"},
									{Kind: yaml.ScalarNode, Value: "content"},
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Value: "application/json"},
											{
												Kind: yaml.MappingNode,
												Content: []*yaml.Node{
													{Kind: yaml.ScalarNode, Value: "schema"},
													{
														Kind: yaml.MappingNode,
														Content: []*yaml.Node{
															{Kind: yaml.ScalarNode, Value: "$ref"},
															{Kind: yaml.ScalarNode, Value: errorRef},
														},
													},
												},
											},
										},
									},
								},
							},
							{Kind: yaml.ScalarNode, Value: "404", Style: yaml.SingleQuotedStyle},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{Kind: yaml.ScalarNode, Value: "description"},
									{Kind: yaml.ScalarNode, Value: "Not found"},
									{Kind: yaml.ScalarNode, Value: "content"},
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Value: "application/json"},
											{
												Kind: yaml.MappingNode,
												Content: []*yaml.Node{
													{Kind: yaml.ScalarNode, Value: "schema"},
													{
														Kind: yaml.MappingNode,
														Content: []*yaml.Node{
															{Kind: yaml.ScalarNode, Value: "$ref"},
															{Kind: yaml.ScalarNode, Value: errorRef},
														},
													},
												},
											},
										},
									},
								},
							},
							{Kind: yaml.ScalarNode, Value: "500", Style: yaml.SingleQuotedStyle},
							{
								Kind: yaml.MappingNode,
								Content: []*yaml.Node{
									{Kind: yaml.ScalarNode, Value: "description"},
									{Kind: yaml.ScalarNode, Value: "Internal server error"},
									{Kind: yaml.ScalarNode, Value: "content"},
									{
										Kind: yaml.MappingNode,
										Content: []*yaml.Node{
											{Kind: yaml.ScalarNode, Value: "application/json"},
											{
												Kind: yaml.MappingNode,
												Content: []*yaml.Node{
													{Kind: yaml.ScalarNode, Value: "schema"},
													{
														Kind: yaml.MappingNode,
														Content: []*yaml.Node{
															{Kind: yaml.ScalarNode, Value: "$ref"},
															{Kind: yaml.ScalarNode, Value: errorRef},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func camelCase(name string) string {
	if len(name) == 0 {
		return name
	}
	runes := []rune(name)
	runes[0] = []rune(string(runes[0]))[0] + ('a' - 'A')
	if runes[0] < 'a' || runes[0] > 'z' {
		runes[0] = runes[0] - ('a' - 'A')
	}
	return string(runes)
}
