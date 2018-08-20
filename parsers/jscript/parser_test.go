package jscript

import (
	"reflect"
	"testing"
)

func newString(s string) *string {
	return &s
}

func TestParseObject(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Structure
		wantErr bool
	}{
		{
			"simple",
			args{
				[]byte(`{p1:"P1"}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString("P1"),
						},
					},
				},
			},
			false,
		},
		{
			"simple UTF8",
			args{
				[]byte(`{p1:"Hello, 世界"}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString("Hello, 世界"),
						},
					},
				},
			},
			false,
		},
		{
			"simple with apos",
			args{
				[]byte(`{p1:"Hello O'Brian"}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString("Hello O'Brian"),
						},
					},
				},
			},
			false,
		},
		{
			"simple with quote",
			args{
				[]byte(`{p1:'Hello "World"'}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString("Hello \"World\""),
						},
					},
				},
			},
			false,
		},
		{
			`simple with \u00e9`,
			args{
				[]byte(`{p1:'Hello Caf\u00e9'}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString(`Hello Café`),
						},
					},
				},
			},
			false,
		},
		{
			"error",
			args{
				[]byte(`{p1"P1"}`),
			},
			nil,
			true,
		},
		{
			"simple 2 properties",
			args{
				[]byte(`{p1:"P1",p2:'P2'}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString("P1"),
						},
					},
					{
						"p2",
						&Value{
							Str: newString("P2"),
						},
					}},
			},
			false,
		},
		{
			"strings with quote",
			args{
				[]byte(`{p1:"\"P1\"",p2:'\'P2\''}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Str: newString(`"P1"`),
						},
					},
					{
						"p2",
						&Value{
							Str: newString(`'P2'`),
						},
					},
				},
			},
			false,
		},
		{
			"simple array",
			args{
				[]byte(`{p1:["A","B","C"]}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Ar: []*Value{
								{
									Str: newString("A"),
								},
								{
									Str: newString("B"),
								},
								{
									Str: newString("C"),
								},
							},
						},
					},
				},
			},
			false,
		},
		{
			"nested structure",
			args{
				[]byte(`{p1:{p2:"P2"}}`),
			},
			&Structure{
				[]*Property{
					{
						"p1",
						&Value{
							Struct: &Structure{
								[]*Property{
									{
										"p2",
										&Value{
											Str: newString("P2"),
										},
									},
								},
							},
						},
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseObject(tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseObject() = %v, want %v", got, tt.want)
			}
		})
	}
}
