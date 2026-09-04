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
          Her puan, AB Kozmetik Tüzüğü 1223/2009&apos;un Eklerinden{' '}
          <strong>türetilir</strong>: bir madde yasaklıysa, kısıtlıysa ya da
          etikette beyanı zorunlu bir koku alerjeniyse, puanı bunu yansıtır.
          Elle atadığımız hiçbir sayı yok — her içeriğin sayfasında{' '}
          <em>&ldquo;Neden bu puan?&rdquo;</em> bağlantısı puanı üreten kuralı
          ve mevzuat maddesini gösterir. Düşük olan daha iyidir. Yüksek bir
          sayı, içeriğin bitmiş üründeki oranında tehlikeli olduğu anlamına
          gelmez — mevzuatın onu kısıtladığı anlamına gelir.
        </p>

        <table className="w-full mt-4 text-sm border border-brand-100 rounded-lg overflow-hidden">
          <caption className="sr-only">Puanlama rubriği, sürüm 1</caption>
          <thead className="bg-brand-50 text-left">
            <tr>
              <th scope="col" className="px-3 py-2 font-semibold">Durum</th>
              <th scope="col" className="px-3 py-2 font-semibold">Puan</th>
            </tr>
          </thead>
          <tbody className="text-espresso/80">
            {[
              ['Ek II — AB’de kozmetikte kullanımı yasak', '10'],
              ['Ek III — etikette beyanı zorunlu koku alerjeni', '7'],
              ['Ek III — konsantrasyon sınırlı', '5'],
              ['Ek V / VI — koşullu izinli koruyucu veya UV filtresi', '4'],
              ['Ek IV — izinli renklendirici', '3'],
              ['Eklerde kısıtlama kaydı yok', '2'],
              ['SCCS olumsuz görüş bildirdi', '+2'],
            ].map(([condition, score]) => (
              <tr key={condition} className="border-t border-brand-100">
                <td className="px-3 py-2">{condition}</td>
                <td className="px-3 py-2 font-semibold tabular-nums">{score}</td>
              </tr>
            ))}
          </tbody>
        </table>

        <p className="text-espresso/80 leading-relaxed mt-4">
          Mevzuatta kaydı olmayan içerikler <strong>puansız kalır</strong> ve
          &ldquo;henüz puanlanmadı&rdquo; olarak görünür. Dayanağı olmayan bir
          sayı uydurmaktansa boş bırakmayı tercih ediyoruz: sıfır göstermek, o
          içeriği ölçeğin en güvenli ucuna yerleştirmek olurdu.
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
