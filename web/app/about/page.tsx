// app/about/page.tsx
import Link from 'next/link'

export const metadata = {
  title: 'Hakkında · Clarity',
}

export default function AboutPage() {
  return (
    <div className="max-w-2xl mx-auto py-12 space-y-8">
      <div>
        <h1 className="text-4xl font-bold mb-4">Clarity hakkında</h1>
        <p className="text-espresso/80 leading-relaxed">
          Clarity, bir makyaj ürününü içeriklerine ayırır, her birini puanlar,
          senin tepki verdiklerini işaretler ve benzer içerik profiline sahip
          daha uygun fiyatlı ürünleri arar. Adı da buradan geliyor: etiketin
          arkasında ne olduğunu netleştirmek.
        </p>
      </div>

      <section>
        <h2 className="text-2xl font-bold mb-3">Puanlar nasıl çalışır?</h2>
        <p className="text-espresso/80 leading-relaxed">
          Her içerik 1 ile 10 arasında bir risk seviyesi taşır; bu, EWG Skin
          Deep veritabanının kullandığı geleneği izler. Düşük olan daha iyidir.
          Yüksek bir sayı, o içeriğin bitmiş bir üründeki oranında tehlikeli
          olduğu anlamına gelmez — bilinmeye değer olduğu anlamına gelir.
        </p>
      </section>

      <section>
        <h2 className="text-2xl font-bold mb-3">Muadiller nasıl bulunuyor?</h2>
        <p className="text-espresso/80 leading-relaxed">
          İçerik listelerini doğrudan karşılaştırıyoruz: iki ürün, ortak
          içeriklerinin birleşik içerik kümesine oranına göre puanlanır. Aynı
          kategorideki yakın eşleşmeler “muadil”, daha gevşek eşleşmeler
          “alternatif” olarak etiketlenir. Kayıtlı bir profilin varsa,
          alerjenlerini içeren her şey önce elenir.
        </p>
      </section>

      <section className="p-6 bg-clay/10 border border-clay/40 rounded-xl">
        <h2 className="text-xl font-bold mb-2 text-cocoa">Tıbbi tavsiye değildir</h2>
        <p className="text-cocoa text-sm leading-relaxed">
          Bu araç içerik etiketlerini okumana yardım eder; bir alerjiyi teşhis
          edemez. İçerik verileri eksik veya güncelliğini yitirmiş olabilir ve
          formülasyonlar değişir. Yeni ürünleri her zaman küçük bir alanda
          deneyin, satın aldığınız ambalajın kendi etiketini kontrol edin ve
          geçmeyen reaksiyonlar için bir dermatoloğa görünün.
        </p>
      </section>

      <section>
        <h2 className="text-2xl font-bold mb-3">Şu anki durum</h2>
        <p className="text-espresso/80 leading-relaxed mb-4">
          Bu bir Faz 1 MVP&apos;si. Katalog, kurgusal marka adları altında
          örnek verilerle dolduruldu ve fotoğraf analizi ürünleri hâlâ elle
          eşleştiriyor — görüntü tanıma Faz 2&apos;de geliyor.
        </p>
        <Link
          href="/ingredients"
          className="text-brand-400 hover:text-brand-500 hover:underline"
        >
          İçerik kataloğuna göz at →
        </Link>
      </section>
    </div>
  )
}
