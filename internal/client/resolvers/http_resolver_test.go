package resolvers

import "testing"

func Test_repair(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no need",
			args: args{
				body: "nonewline",
			},
			want: "nonewline",
		},
		{
			name: "replace newline",
			args: args{
				body: "nonewline\n",
			},
			want: "nonewline",
		},
		{
			name: "replace newline start",
			args: args{
				body: "\nnonewline",
			},
			want: "nonewline",
		},
		{
			name: "replace newlines",
			args: args{
				body: "\nnonewline\n",
			},
			want: "nonewline",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repair(tt.args.body); got != tt.want {
				t.Errorf("repair() = %v, want %v", got, tt.want)
			}
		})
	}
}
