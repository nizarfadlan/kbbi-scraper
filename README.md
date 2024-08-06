# Scraper KBBI
Scraper KBBI 5 pada website [kbbi.kemdikbud](https://kbbi.kemdikbud.go.id/) dengan jumlah daftar kata 112.651.

# Example data

Dalam penyimpanan data 1 kata bisa lebih dari 1 lema dan 1 lema bisa lebih dari 1 arti (terdiri dari kelas kata dan keterangan). Ada juga kata yang tidak memiliki kelas kata.

|id |kata      |lema         |kelas_kata                                            |keterangan                                                                                             |
|---|----------|-------------|------------------------------------------------------|-------------------------------------------------------------------------------------------------------|
|33 |alas cawan|alas cawan   |                                                      |lapik cangkir                                                                                          |
|---|---       |---          |---                                                   |---                                                                                                    |
|57 |ankilosis |an.ki.lo.sis |n[Nomina: kata benda] Dok[Kedokteran dan Fisiologi: -]|tergabungnya tulang-tulang atau bagian lain yang keras dan membentuk satu tulang atau bagian yang keras|
|58 |ankilosis |an.ki.lo.sis |n[Nomina: kata benda] Dok[Kedokteran dan Fisiologi: -]|kekakuan sendi karena penyakit atau pembedahan                                                         |
|59 |anihilasi |a.ni.hi.la.si|n[Nomina: kata benda]                                 |keadaan hancur atau musnah total                                                                       |
|60 |anihilasi |a.ni.hi.la.si|n[Nomina: kata benda] Fis[Fisika: -]                  |kombinasi dari sebuah partikel dan antipartikel seperti elektron dan positron untuk menghasilkan energi|
|...|...       |...          |...                                                   |...                                                                                                    |


# Source

Kumpulan kata didapat dari repository [damzaky/kumpulan-kata-bahasa-indonesia-KBBI](https://github.com/damzaky/kumpulan-kata-bahasa-indonesia-KBBI)
