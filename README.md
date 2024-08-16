# Scraper KBBI
Scraper KBBI 5 pada website [kbbi.kemdikbud](https://kbbi.kemdikbud.go.id/) dengan jumlah daftar kata 112.651.

# Run

```bash
go run main.go
```

Atau build terlebih dahulu

```bash
go build

./kbbi-scraper

# or

kbbi-scraper.exe
```

# Example data

Dalam penyimpanan data 1 kata bisa lebih dari 1 lema dan 1 lema bisa lebih dari 1 arti (terdiri dari kelas kata dan keterangan). Ada juga kata yang tidak memiliki kelas kata.

|id |kata      |lema         |kelas_kata                                            |keterangan                                                                                             |
|---|----------|-------------|------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
|33 |alas cawan|alas cawan   |                                                      |lapik cangkir                                                                                          |
|---|---       |---          |---                                                   |---                                                                                                    |
|93 |bagaimana |ba.gai.ma.na bentuk tidak baku: begimana, gimana|pron[Pronomina: kelas kata yang meliputi kata ganti, kata tunjuk, dan kata tanya]|kata tanya untuk menanyakan cara, perbuatan (lazimnya diikuti kata cara): -- caranya membeli buku dari luar negeri?|
|94 |bagaimana |ba.gai.ma.na bentuk tidak baku: begimana, gimana|pron[Pronomina: kelas kata yang meliputi kata ganti, kata tunjuk, dan kata tanya]|kata tanya untuk menanyakan akibat suatu tindakan: -- kalau dia lari nanti?                            |
|95 |bagaimana |ba.gai.ma.na bentuk tidak baku: begimana, gimana|pron[Pronomina: kelas kata yang meliputi kata ganti, kata tunjuk, dan kata tanya]|kata tanya untuk meminta pendapat dari kawan bicara (diikuti kata kalau): -- kalau kita pergi ke Puncak?|
|96 |bagaimana |ba.gai.ma.na bentuk tidak baku: begimana, gimana|pron[Pronomina: kelas kata yang meliputi kata ganti, kata tunjuk, dan kata tanya]|kata tanya untuk menanyakan penilaian atas suatu gagasan: -- pendapatmu?                               |
|---|---       |---          |---                                                   |---                                                                                                    |
|1286|aku       |a.ku         |pron[Pronomina: kelas kata yang meliputi kata ganti, kata tunjuk, dan kata tanya]|kata ganti orang pertama yang berbicara atau yang menulis (dalam ragam akrab); diri sendiri; saya      |
|1287|aku       |a.ku         |n[Nomina: kata benda] akr[akronim]                    |anggaran dan keuangan                                                                                  |
|1288|aku       |A.ku         |n[Nomina: kata benda]                                 |Li'o                                                                                                   |
|...|...       |...          |...                                                   |...                                                                                                    |

# Note

Kekurangan masih belum bisa mengembail kata prakategorial (contoh: https://kbbi.kemdikbud.go.id/entri/repuh), jika menemukan kata prakategorial akan dianggap tidak ada hasil yang nantinya masuk ke file `no_result_word.json`

# Source

Kumpulan kata didapat dari repository [damzaky/kumpulan-kata-bahasa-indonesia-KBBI](https://github.com/damzaky/kumpulan-kata-bahasa-indonesia-KBBI)
