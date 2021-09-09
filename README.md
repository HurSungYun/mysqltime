# mysqltime

## motivation

MySQL has [the TIME type](https://dev.mysql.com/doc/refman/8.0/en/time.html) which supports saving times.

According to its spec, it is said that 

```
TIME values may range from '-838:59:59' to '838:59:59'.
The hours part may be so large because the TIME type can be used not only to represent a time of day (which must be less than 24 hours),
but also elapsed time or a time interval between two events (which may be much greater than 24 hours, or even negative).
```

Because of its spec, it is hard to unmarshal to primitive golang data structures.

https://github.com/go-sql-driver/mysql/issues/849

It would be really great to support custom data type `mysqltime.Time` to support use cases for mysql TIME type.

It can be either `time of day` or `duration`.

## interface

```
type Time interface {
    TimeOfDay() (int64, error) // maybe?
    Duration() time.Duration
}
```

