# test-temtera
Technical test for Back-end at Temtera

How to run:
1) Make sure you have docker installed on your system
2) Change directory terminal/bash to the root of this repo
3) Run `docker-compose up` and enjoy

Endpoints:
1) Create product
```
POST localhost:5000/product

Payload:
{
    "product_name": "string",
    "price": 100,
    "qty": 1,
    "category": "string"
}
```
2) Find products with query param 
```
GET localhost:5001/product?query=string
```

Note: if you want to change the code and re-build the images, run `docker-compose up --build`