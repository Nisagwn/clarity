// app/products/[id]/page.tsx — içerikler ve muadillerle ürün detayı
import Link from "next/link";
import { notFound } from "next/navigation";
import IngredientCard from "@/components/IngredientCard";
import ProductImage from "@/components/ProductImage";
import {
  getProduct,
  getDupes,
  formatPrice,
  dataSource,
  ApiError,
} from "@/lib/api";
import type { Product, Recommendation } from "@/lib/api";

export default async function ProductDetailPage({
  params,
}: {
  params: { id: string };
}) {
  let product: Product;
  try {
    product = await getProduct(params.id);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
  }

  // Muadiller güzel bir ek; buradaki bir hata sayfayı bozmamalı.
  let dupes: Recommendation[] = [];
  try {
    dupes = (await getDupes(params.id, 4)).recommendations;
  } catch {
    dupes = [];
  }

  const ingredients = product.ingredients ?? [];
  // Puansız içerik burada sayılmaz: bilinmeyen bir puan "yüksek risk" değil.
  const riskiest = ingredients.filter((i) => (i.concern_level ?? 0) >= 7);
  const withAllergens = ingredients.filter((i) => i.allergens.length > 0);

  return (
    <div className="max-w-4xl mx-auto py-8">
      <Link
        href="/products"
        className="text-brand-400 hover:text-brand-500 hover:underline text-sm"
      >
        ← Ürünlere dön
      </Link>

      <header className="mt-4 mb-8 flex flex-col sm:flex-row gap-6">
        <ProductImage
          src={product.image_url}
          alt={`${product.brand} ${product.name}`}
          size="detail"
        />

        <div className="min-w-0">
          <p className="text-sm uppercase tracking-wide text-mauve">
            {product.brand}
          </p>
          <h1 className="text-4xl font-bold mt-1">{product.name}</h1>
          <p className="text-lg font-semibold mt-2">
            {formatPrice(product.price, product.currency)}
            {product.category && (
              <span className="font-normal text-ink-muted">
                {" "}
                · {product.category}
              </span>
            )}
          </p>
          {product.description && (
            <p className="text-espresso/80 mt-3">{product.description}</p>
          )}

          {product.data_quality === "incomplete" && (
            <p className="mt-4 p-3 text-sm bg-clay/10 border border-clay/40 rounded-lg text-cocoa">
              <span className="font-semibold">İçerik listesi eksik. </span>
              Bu üründe üçten az içerik kayıtlı; liste tamamlanana kadar muadil
              karşılaştırmasına alınmıyor. Eksik listeyle yapılan karşılaştırma,
              benzerlik değil veri boşluğu ölçer.
            </p>
          )}
        </div>
      </header>

      {/* Bir bakışta özet */}
      <div className="grid gap-4 sm:grid-cols-3 mb-10">
        <Stat label="İçerik sayısı" value={String(ingredients.length)} />
        <Stat
          label="Alerjen taşıyan"
          value={String(withAllergens.length)}
          tone={withAllergens.length > 0 ? "clay" : "sage"}
        />
        <Stat
          label="Yüksek riskli (7+)"
          value={String(riskiest.length)}
          tone={riskiest.length > 0 ? "brand" : "sage"}
        />
      </div>

      <section className="mb-12">
        <h2 className="text-2xl font-bold mb-4">İçerik dökümü</h2>
        <p className="text-sm text-ink-muted mb-4">
          INCI sırasına göre listelenir — üstteki içerikler en yüksek oranda
          bulunanlardır.
        </p>
        {ingredients.length === 0 ? (
          <p className="text-mauve">Bu ürüne henüz içerik eşlenmemiş.</p>
        ) : (
          <div className="space-y-3">
            {ingredients.map((ing) => (
              <IngredientCard key={ing.id} ingredient={ing} />
            ))}
          </div>
        )}
      </section>

      {dupes.length > 0 && (
        <section>
          <h2 className="text-2xl font-bold mb-2">
            Muadiller ve alternatifler
          </h2>
          <p className="text-sm text-ink-muted mb-4">
            Bu ürünle içerik listesini ne kadar paylaştıklarına göre sıralandı.
          </p>
          <div className="grid gap-4 sm:grid-cols-2">
            {dupes.map((rec) => (
              <Link
                key={rec.id}
                href={`/products/${rec.id}`}
                className="border border-brand-100 rounded-xl p-5 bg-white hover:border-brand-300 transition"
              >
                <div className="flex items-center justify-between mb-2">
                  <span
                    className={`text-xs font-semibold px-2 py-1 rounded ${
                      rec.type === "dupe"
                        ? "bg-brand-100 text-brand-500"
                        : "bg-mist/25 text-mauve"
                    }`}
                  >
                    {rec.type === "dupe" ? "muadil" : "alternatif"}
                  </span>
                  <span className="text-sm font-semibold">
                    %{Math.round(rec.similarity_score * 100)} eşleşme
                  </span>
                </div>
                <p className="text-xs uppercase tracking-wide text-mauve">
                  {rec.brand}
                </p>
                <h3 className="font-semibold">{rec.name}</h3>
                <p className="text-sm font-semibold mt-1">
                  {formatPrice(rec.price, rec.currency)}
                  {rec.price !== null &&
                    product.price !== null &&
                    rec.price < product.price && (
                      <span className="text-sage font-normal">
                        {" "}
                        · {formatPrice(
                          product.price - rec.price,
                          rec.currency,
                        )}{" "}
                        tasarruf
                      </span>
                    )}
                </p>
                <p className="text-sm text-espresso/70 mt-2">{rec.reason}</p>
              </Link>
            ))}
          </div>
        </section>
      )}

      <Attribution product={product} />
    </div>
  );
}

/**
 * Kaynak atfı. ODbL "atıf" ve "aynı lisansla paylaşma" gerektiriyor, ürün
 * görselleri de CC-BY-SA; veriyi gösterip kaynağı göstermemek lisans ihlali.
 * Bu yüzden bileşen koşullu değil: kaynağı olan her ürün sayfasında görünür.
 */
function Attribution({ product }: { product: Product }) {
  const source = dataSource(product.source);
  if (!source) return null;

  return (
    <footer className="mt-12 pt-4 border-t border-brand-100 text-xs text-ink-muted space-y-1">
      <p>
        Veri:{" "}
        <a
          href={product.source_url ?? source.url}
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand-400 hover:text-brand-500 hover:underline"
        >
          {source.label}
        </a>{" "}
        ·{" "}
        <a
          href={source.licenseUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="text-brand-400 hover:text-brand-500 hover:underline"
        >
          {source.license}
        </a>
        {product.image_url && " · ürün görseli CC-BY-SA"}
      </p>
      {product.verified_at && (
        <p>
          Kaynaktan son alınma:{" "}
          {new Date(product.verified_at).toLocaleDateString("tr-TR")}
        </p>
      )}
      <p>
        İçerik listeleri topluluk katkısıyla derlenir ve eksik ya da
        güncelliğini yitirmiş olabilir. Satın aldığınız ambalajın kendi etiketi
        esastır.
      </p>
    </footer>
  );
}

function Stat({
  label,
  value,
  tone = "neutral",
}: {
  label: string;
  value: string;
  tone?: "neutral" | "sage" | "clay" | "brand";
}) {
  const tones = {
    neutral: "bg-white border-brand-100",
    sage: "bg-sage/10 border-sage/30",
    clay: "bg-clay/15 border-clay/40",
    brand: "bg-brand-100 border-brand-300",
  };

  return (
    <div className={`border rounded-xl p-4 ${tones[tone]}`}>
      <p className="text-2xl font-bold">{value}</p>
      <p className="text-sm text-espresso/70">{label}</p>
    </div>
  );
}
