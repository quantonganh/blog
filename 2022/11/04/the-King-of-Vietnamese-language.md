---
title: The King of Vietnamese language game show
date: Thu Nov 04 10:12:50 +07 2022
description:
categories:
- Programming
tags:
- golang
---

My family likes to watch "The King of Vietnamese language" game show together.
I want to encourage our son to love Vietnamese.
At the end of the game, the player has to find 7 complex words from the letters, for e.g,

> đ / ă / n / g / c / a / y

One evening a few weeks ago, while we were watching the final round, 
my wife suddenly came up with an idea: this game could be programmed.

As soon as the game show ends, I immediately sat down at my computer and [started coding](https://github.com/quantonganh/vtv).

The first algorithm that comes to my mind is exactly the way I would solve it on paper:

- separate vowels and consonants
- combine consonant and vowel / vowel and consonant
- find simple words
- find complex words

At the final step, the valid words can be filtered out by using a [Vietnamese word list](http://www.informatik.uni-leipzig.de/~duc/software/misc/wordlist.html).

Here's [the result](/vtv):

```
$ vtv "đăngcay"
      à à|   a đảng|    à này|       á à|     ay áy|   áy náy|      ăn ý|  ăn năn|     ăn cá|   ăng ẳng|
   ằng ặc|    ặc ặc|  ác đảng|      ác ý|     đa đa|  đa đảng|   đa năng|   đá gà|    đay đả|   đay đảy|
 đăng đàn| đằng ngà|    đắc ý| đằng đằng| đằng đẵng| đằng nay|  đằng này|  đàn đá|   đàn đáy|  đàng này|
  đàn gảy|   đàn ca|    na ná|   này này|  nằng nặc|  ngà ngà| ngày ngày| ngay cả| ngay ngáy| ngày đàng|
 ngày đản| ngày nay| ngày này| ngày càng|  ngăn cản|    gà gà|    gà gáy|   gà ác|     gà đá|    gay ác|
  gay gay|  gằn gằn|  gặc gặc|   gàn gàn|     ca ca|    cà na|    cá gáy| cá căng|     cả cả|   cay cảy|
  cày cạy|  cày ngả| càng cạc|   cạc cạc|
```

One thing that I've noticed is that players tend to look for complex words that start with consonants, 
but forget words that start with a single vowel.

Hope that I can join this game in the future.