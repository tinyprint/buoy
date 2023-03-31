# buoy

make your local dev a lot more production

## usage

```
# get buoy ready for real life
$ sudo buoy -setup

# forward www.b.com to localhost:2222
# forward www.b.com/blog to localhost:3333
# forward app.b.com to localhost:4444
$ buoy www:2222 www/blog:3333 app:4444
```
