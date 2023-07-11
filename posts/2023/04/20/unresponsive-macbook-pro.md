---
title: Load averages approximately 1000 on macOS?
date: 2023-04-20
categories:
- Development Environment
tags:
- macOS
description:
---
Em "vợ" ra đời năm 2015, cưới về năm 2018, loa đã rè, pin đã chai.

Một ngày đẹp trời bật máy lên và không làm được gì. Load gần... 1000 (trong ảnh là sau khi đã restart, và hơn 30p sau mới gõ được `w`):

![Load Averages 600](/2023/04/20/load-averages-600.jpeg)

Mình thử khởi động lại thì mất khoảng 5-10p mới đến chỗ gõ password.
Search thấy có bạn [thay cable ổ cứng thì hết](https://discussions.apple.com/thread/8308724) nhưng đời 2015 thì ổ cứng không có cable.

Mình thử chạy Disk Utility thì báo "corrupt":

![Macintosh HD corrupt](/2023/04/20/macintosh-hd-corrupt.jpeg)

Điều lạ là mình boot từ ổ cứng gắn ngoài vẫn chậm.

Mình thử cài lại thì đến "Less than a minute remaining..." và sáng hôm sau ngủ dậy vẫn ở đó:

![Less than a minute remaining](/2023/04/20/less-than-a-minute-remaining.jpeg)

---
Mình cầm máy qua một chỗ thì các bạn ấy chưa tìm ra nguyên nhân.
Gúc tiếp thì tìm thấy [UFOTech](https://voz.vn/t/nhan-sua-laptop-macboock-cho-ae-voz-o-hn.420922/)
Và bạn ấy đã sửa được.

Thứ nhất là hỏng ét-đi (ssd)\
Thứ nhì là hỏng cái gì trên main

Bạn ấy chỉ nói là có cái gì đó trên main bị phù (do ẩm).

Bạn nào cần tìm chỗ sửa MacBook thì cân nhắc ở đây nhé, mình thấy bạn ấy làm có vẻ ổn.

PS: Hình như trên server mình cũng chưa bao giờ thấy load cao thế này.
"Uninterruptible Sleep" thì đã từng gặp nhưng không nhớ load averages là bao nhiêu.