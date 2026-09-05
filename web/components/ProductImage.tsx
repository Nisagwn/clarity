// components/ProductImage.tsx — ürün görseli, görsel yoksa yer tutucu
import Image from 'next/image'

/**
 * Katalog görselleri Open Beauty Facts'ten geliyor ve CC-BY-SA lisanslı;
 * atıf ürün sayfasının altındaki kaynak bloğunda veriliyor.
 *
 * Görseli olmayan ürün gizlenmez: kataloğun büyük bölümünde fotoğraf yok ve
 * asıl bilgi içerik listesinde. Yer tutucu, kartların hizasını bozmadan
 * "fotoğraf yok" der.
 */
export default function ProductImage({
  src,
  alt,
  size = 'card',
}: {
  src?: string
  alt: string
  size?: 'card' | 'detail'
}) {
  const box =
    size === 'detail'
      ? 'w-40 h-40 sm:w-52 sm:h-52'
      : 'w-20 h-20'

  if (!src) {
    return (
      <div
        className={`${box} shrink-0 rounded-lg bg-brand-50 border border-brand-100 flex items-center justify-center text-[10px] text-ink-muted text-center px-2`}
        aria-hidden="true"
      >
        fotoğraf yok
      </div>
    )
  }

  return (
    <div
      className={`${box} shrink-0 relative rounded-lg overflow-hidden bg-brand-50 border border-brand-100`}
    >
      <Image
        src={src}
        alt={alt}
        fill
        sizes={size === 'detail' ? '208px' : '80px'}
        className="object-contain"
        // Katalogdaki adresler topluluk verisinden geliyor; ölü bağlantı
        // sayfayı bozmamalı.
        unoptimized
      />
    </div>
  )
}
