---
title: Automate office tasks with Python
date: 2021-06-10
description: Lập trình có gì hay?
categories:
- Programming
tags:
- python
- pandas
---
Hàng tháng bạn nhận được một file excel về danh sách D. Từ danh sách này bạn cần filter cột C theo tiêu chí T1 để lấy ra danh sách D1. Có được D1 rồi bạn Save As thành t1.xlsx.
Làm xong T1, bạn tiếp tục với T2 -> t2.xlsx
...
Làm tiếp cho đến Tn -> tn.xlsx.
Có được các danh sách t1.xlsx -> tn.xlsx này rồi, giờ bạn cần gửi mail:

- t1.xlsx: gửi cho P1, P2, P3
- t2.xlsx: gửi cho P4, P5
- ...
  
Tháng nào cũng mất nguyên buổi sáng (có khi không xong) để ngồi làm mấy việc này. Hỏi có chán không? Chán chứ.
Đây là lúc lập trình có thể giúp bạn.
Mình đề nghị:

- Bắt đầu với CS50: https://www.youtube.com/channel/UCcabW7890RKJzL968QWEykA
- Học kỹ bài Python: https://www.youtube.com/watch?v=ZEQh45W_UDo
- Đọc về pandas để làm việc với file excel: https://pandas.pydata.org/
- Đọc Win32 để gọi Outlook: https://pypi.org/project/pywin32/

Sau đó thì lập một tài khoản GitHub rồi code thôi. Loanh quanh đâu đó khoảng 15, 20 dòng là xong.
  Giờ thì bạn chạy chương trình do mình viết ra, nó tự export ra các file excel cho bạn, rồi gửi mail đi. Hỏi có thích không? Thích chứ! (Vì sau đó bạn có nguyên buổi sáng để ngồi lướt fb 😃).