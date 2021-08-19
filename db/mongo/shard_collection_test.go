package mongo

import (
	"reflect"
	"testing"
	"time"
)

func TestShardCollection_calcDaySpan(t *testing.T) {
	type fields struct {
		prefix  string
		daySpan map[int]int
	}
	type args struct {
		day int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[int]int
	}{
		{
			name: "31day 1span",
			fields: fields{
				prefix:  "test",
				daySpan: make(map[int]int),
			},
			args: args{day: 31},
			want: map[int]int{
				1:  1,
				2:  1,
				3:  1,
				4:  1,
				5:  1,
				6:  1,
				7:  1,
				8:  1,
				9:  1,
				10: 1,
				11: 1,
				12: 1,
				13: 1,
				14: 1,
				15: 1,
				16: 1,
				17: 1,
				18: 1,
				19: 1,
				20: 1,
				21: 1,
				22: 1,
				23: 1,
				24: 1,
				25: 1,
				26: 1,
				27: 1,
				28: 1,
				29: 1,
				30: 1,
				31: 1,
			},
		},
		{
			name: "16day 2_span",
			fields: fields{
				prefix:  "test",
				daySpan: make(map[int]int),
			},
			args: args{day: 16},
			want: map[int]int{
				1:  1,
				2:  1,
				3:  1,
				4:  1,
				5:  1,
				6:  1,
				7:  1,
				8:  1,
				9:  1,
				10: 1,
				11: 1,
				12: 1,
				13: 1,
				14: 1,
				15: 1,
				16: 1,
				17: 2,
				18: 2,
				19: 2,
				20: 2,
				21: 2,
				22: 2,
				23: 2,
				24: 2,
				25: 2,
				26: 2,
				27: 2,
				28: 2,
				29: 2,
				30: 2,
				31: 2,
			},
		},
		{
			name: "7day 5_span",
			fields: fields{
				prefix:  "test",
				daySpan: make(map[int]int),
			},
			args: args{day: 7},
			want: map[int]int{
				1:  1,
				2:  1,
				3:  1,
				4:  1,
				5:  1,
				6:  1,
				7:  1,
				8:  2,
				9:  2,
				10: 2,
				11: 2,
				12: 2,
				13: 2,
				14: 2,
				15: 3,
				16: 3,
				17: 3,
				18: 3,
				19: 3,
				20: 3,
				21: 3,
				22: 4,
				23: 4,
				24: 4,
				25: 4,
				26: 4,
				27: 4,
				28: 4,
				29: 5,
				30: 5,
				31: 5,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coll := ShardCollection{
				Prefix:  tt.fields.prefix,
				daySpan: tt.fields.daySpan,
			}
			if got := coll.calcDaySpan(tt.args.day); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calcDaySpan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShardCollection_EncodeCollName(t *testing.T) {
	type fields struct {
		prefix  string
		daySpan map[int]int
	}
	type args struct {
		day       int
		bucket    string
		timestamp int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "7day 5span",
			fields: fields{
				prefix:  "test",
				daySpan: make(map[int]int),
			},
			args: args{day: 7, bucket: "1", timestamp: time.Now().Unix()},
			want: "test_1_202104_05",
		},
		{
			name: "16day 2span",
			fields: fields{
				prefix:  "test",
				daySpan: make(map[int]int),
			},
			args: args{day: 16, bucket: "1", timestamp: time.Now().Unix()},
			want: "test_1_202104_02",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coll := ShardCollection{
				Prefix:  tt.fields.prefix,
				daySpan: tt.fields.daySpan,
			}
			coll.daySpan = coll.calcDaySpan(tt.args.day)
			if got := coll.EncodeCollName(tt.args.bucket, tt.args.timestamp); got != tt.want {
				t.Errorf("EncodeCollName() = %v, want %v", got, tt.want)
			}
		})
	}
}
