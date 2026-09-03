// app/profile/page.tsx — hesap, cilt profili ve rıza yönetimi
'use client'

import { useCallback, useEffect, useState } from 'react'
import Link from 'next/link'
import {
  register,
  login,
  logout,
  getMyProfile,
  updateMyProfile,
  updateConsent,
  deleteAccount,
  SKIN_TYPES,
  skinTypeLabel,
  ApiError,
  type UserProfile,
} from '@/lib/api'

type Mode = 'register' | 'login'

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [restoring, setRestoring] = useState(true)
  const [mode, setMode] = useState<Mode>('register')

  const refresh = useCallback(async () => {
    try {
      setProfile(await getMyProfile())
    } catch {
      // 401 beklenen durum: henüz giriş yapılmamış.
      setProfile(null)
    } finally {
      setRestoring(false)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  if (restoring) {
    return <p className="py-12 text-center text-ink-muted">Yükleniyor…</p>
  }

  if (profile) {
    return <ProfileView profile={profile} onChange={refresh} />
  }

  return <AuthForms mode={mode} setMode={setMode} onSuccess={refresh} />
}

// ===== Giriş / kayıt =====

function AuthForms({
  mode,
  setMode,
  onSuccess,
}: {
  mode: Mode
  setMode: (m: Mode) => void
  onSuccess: () => void
}) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [skinType, setSkinType] = useState('sensitive')
  const [allergens, setAllergens] = useState<string[]>([])
  const [healthConsent, setHealthConsent] = useState(false)
  const [marketingConsent, setMarketingConsent] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setError('')

    try {
      if (mode === 'login') {
        await login(email.trim(), password)
      } else {
        await register({
          email: email.trim(),
          password,
          skin_type: skinType,
          health_data_consent: healthConsent,
          marketing_consent: marketingConsent,
          allergens: healthConsent ? allergens : [],
        })
      }
      onSuccess()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Bir sorun oluştu.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto py-12">
      <h1 className="text-4xl font-bold mb-2">
        {mode === 'register' ? 'Hesap Oluştur' : 'Giriş Yap'}
      </h1>
      <p className="text-espresso/70 mb-8">
        {mode === 'register'
          ? 'Cildinin neye tepki verdiğini kaydet; uygulamanın her yerinde o içerikleri senin için işaretleyelim.'
          : 'Kayıtlı profiline dön.'}
      </p>

      <form
        onSubmit={handleSubmit}
        className="border border-brand-100 rounded-xl p-6 bg-white space-y-5"
      >
        <div>
          <label htmlFor="email" className="block text-sm font-semibold mb-1">
            E-posta
          </label>
          <input
            id="email"
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="sen@ornek.com"
            className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-500"
          />
        </div>

        <div>
          <label htmlFor="password" className="block text-sm font-semibold mb-1">
            Parola
          </label>
          <input
            id="password"
            type="password"
            required
            minLength={mode === 'register' ? 10 : undefined}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-500"
          />
          {mode === 'register' && (
            <p className="text-sm text-ink-muted mt-1">En az 10 karakter.</p>
          )}
        </div>

        {mode === 'register' && (
          <>
            <div>
              <label
                htmlFor="skin_type"
                className="block text-sm font-semibold mb-1"
              >
                Cilt tipi
              </label>
              <select
                id="skin_type"
                value={skinType}
                onChange={(e) => setSkinType(e.target.value)}
                className="w-full px-4 py-2 border border-brand-200 rounded-lg bg-white"
              >
                {SKIN_TYPES.map((t) => (
                  <option key={t} value={t}>
                    {skinTypeLabel(t)}
                  </option>
                ))}
              </select>
            </div>

            <ConsentBlock
              healthConsent={healthConsent}
              setHealthConsent={setHealthConsent}
              marketingConsent={marketingConsent}
              setMarketingConsent={setMarketingConsent}
            />

            {healthConsent && (
              <AllergenInput value={allergens} onChange={setAllergens} />
            )}
          </>
        )}

        {error && (
          <div
            role="alert"
            className="p-4 bg-brand-100 border border-brand-300 text-brand-500 rounded-lg"
          >
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={busy}
          className="w-full py-3 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {busy
            ? 'Gönderiliyor…'
            : mode === 'register'
              ? 'Hesabı oluştur'
              : 'Giriş yap'}
        </button>

        <p className="text-sm text-center text-ink-muted">
          {mode === 'register' ? 'Zaten hesabın var mı? ' : 'Hesabın yok mu? '}
          <button
            type="button"
            onClick={() => {
              setMode(mode === 'register' ? 'login' : 'register')
              setError('')
            }}
            className="text-brand-500 font-semibold hover:underline"
          >
            {mode === 'register' ? 'Giriş yap' : 'Hesap oluştur'}
          </button>
        </p>
      </form>
    </div>
  )
}

// ===== Rıza =====

/**
 * Sağlık verisi ve pazarlama rızaları AYRI sorulur ve ikisi de varsayılan
 * kapalıdır. Formun gönderilmesi hiçbirine bağlı değildir: reddetmek kabul
 * etmek kadar kolay olmalı. Bir sağlık uygulaması, sağlık verisi rızasını
 * pazarlama rızasıyla paketlediği için ceza almıştı.
 */
function ConsentBlock({
  healthConsent,
  setHealthConsent,
  marketingConsent,
  setMarketingConsent,
}: {
  healthConsent: boolean
  setHealthConsent: (v: boolean) => void
  marketingConsent: boolean
  setMarketingConsent: (v: boolean) => void
}) {
  return (
    <fieldset className="border border-brand-200 rounded-lg p-4 space-y-4">
      <legend className="px-2 text-sm font-semibold">İzinler</legend>

      <label className="flex gap-3 items-start cursor-pointer">
        <input
          type="checkbox"
          checked={healthConsent}
          onChange={(e) => setHealthConsent(e.target.checked)}
          className="mt-1 w-5 h-5 shrink-0 accent-brand-500"
        />
        <span className="text-sm">
          <strong className="block">
            Alerjen bilgimin işlenmesine izin veriyorum
          </strong>
          <span className="text-ink-muted">
            Alerjen listen sağlık verisi sayılır. Yalnızca ürünleri senin için
            işaretlemekte kullanılır, üçüncü taraflarla paylaşılmaz. Bu izni
            vermezsen hesabını yine açabilirsin; yalnızca alerjen işaretlemesi
            çalışmaz.{' '}
            <Link href="/about" className="text-brand-500 hover:underline">
              Nasıl kullanıldığını oku
            </Link>
          </span>
        </span>
      </label>

      <label className="flex gap-3 items-start cursor-pointer">
        <input
          type="checkbox"
          checked={marketingConsent}
          onChange={(e) => setMarketingConsent(e.target.checked)}
          className="mt-1 w-5 h-5 shrink-0 accent-brand-500"
        />
        <span className="text-sm">
          <strong className="block">E-posta bülteni almak istiyorum</strong>
          <span className="text-ink-muted">
            Tamamen isteğe bağlı, istediğin an bırakabilirsin.
          </span>
        </span>
      </label>
    </fieldset>
  )
}

// ===== Alerjen girişi (çip tabanlı) =====

const COMMON_ALLERGENS = [
  'parfüm',
  'nikel',
  'lanolin',
  'formaldehit',
  'paraben',
  'sülfat',
]

function AllergenInput({
  value,
  onChange,
}: {
  value: string[]
  onChange: (v: string[]) => void
}) {
  const [draft, setDraft] = useState('')

  const add = (raw: string) => {
    const term = raw.trim().toLowerCase()
    if (term && !value.includes(term)) onChange([...value, term])
    setDraft('')
  }

  return (
    <div>
      <label htmlFor="allergen-input" className="block text-sm font-semibold mb-1">
        Bilinen alerjenlerin
      </label>

      {value.length > 0 && (
        <ul className="flex flex-wrap gap-2 mb-2">
          {value.map((a) => (
            <li
              key={a}
              className="flex items-center gap-1 pl-3 pr-1 py-1 bg-brand-100 text-brand-500 rounded text-sm"
            >
              {a}
              <button
                type="button"
                onClick={() => onChange(value.filter((x) => x !== a))}
                aria-label={`${a} alerjenini kaldır`}
                className="w-6 h-6 flex items-center justify-center rounded hover:bg-brand-200"
              >
                <span aria-hidden="true">×</span>
              </button>
            </li>
          ))}
        </ul>
      )}

      <input
        id="allergen-input"
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ',') {
            e.preventDefault()
            add(draft)
          }
        }}
        onBlur={() => draft && add(draft)}
        placeholder="Yaz ve Enter'a bas"
        className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-500"
      />

      <p className="text-sm text-ink-muted mt-2 mb-1">Sık görülenler:</p>
      <div className="flex flex-wrap gap-2">
        {COMMON_ALLERGENS.filter((a) => !value.includes(a)).map((a) => (
          <button
            key={a}
            type="button"
            onClick={() => add(a)}
            className="px-3 py-1 border border-brand-200 rounded text-sm hover:bg-brand-50"
          >
            + {a}
          </button>
        ))}
      </div>
    </div>
  )
}

// ===== Giriş yapılmış görünüm =====

function ProfileView({
  profile,
  onChange,
}: {
  profile: UserProfile
  onChange: () => void
}) {
  const [editing, setEditing] = useState(false)
  const [allergens, setAllergens] = useState<string[]>(profile.allergens)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')

  const hasHealthData = profile.allergens.length > 0

  const save = async () => {
    setBusy(true)
    setError('')
    try {
      await updateMyProfile({ skin_type: profile.skin_type, allergens })
      setEditing(false)
      onChange()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Kaydedilemedi.')
    } finally {
      setBusy(false)
    }
  }

  const withdraw = async () => {
    if (
      !window.confirm(
        'Alerjen verinizin tamamı silinecek ve ürünler artık sizin için işaretlenmeyecek. Devam edilsin mi?',
      )
    )
      return

    setBusy(true)
    try {
      await updateConsent('health_data', false)
      onChange()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'İşlem başarısız.')
    } finally {
      setBusy(false)
    }
  }

  const removeAccount = async () => {
    if (
      !window.confirm(
        'Hesabınız ve ilişkili tüm veri kalıcı olarak silinecek. Bu geri alınamaz. Devam edilsin mi?',
      )
    )
      return

    setBusy(true)
    try {
      await deleteAccount()
      onChange()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Silinemedi.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div className="max-w-2xl mx-auto py-12">
      <h1 className="text-4xl font-bold mb-2">Profilim</h1>
      <p className="text-espresso/70 mb-8">
        Yükleme ve ürün sayfaları, tepki verdiğiniz içerikleri işaretlemek için
        bunu kullanır.
      </p>

      <div className="border border-brand-100 rounded-xl p-6 bg-white space-y-4">
        <Field label="E-posta" value={profile.email || '—'} />
        <Field
          label="Cilt tipi"
          value={
            profile.skin_type ? skinTypeLabel(profile.skin_type) : 'belirtilmedi'
          }
        />

        <div>
          <p className="text-sm text-ink-muted mb-2">Alerjenler</p>
          {editing ? (
            <>
              <AllergenInput value={allergens} onChange={setAllergens} />
              <div className="flex gap-2 mt-3">
                <button
                  onClick={save}
                  disabled={busy}
                  className="px-4 py-2 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 disabled:opacity-50"
                >
                  Kaydet
                </button>
                <button
                  onClick={() => {
                    setAllergens(profile.allergens)
                    setEditing(false)
                  }}
                  className="px-4 py-2 border border-brand-200 rounded-lg hover:bg-brand-50"
                >
                  Vazgeç
                </button>
              </div>
            </>
          ) : (
            <>
              {profile.allergens.length === 0 ? (
                <p className="text-espresso/70">Kayıtlı alerjen yok</p>
              ) : (
                <ul className="flex flex-wrap gap-2">
                  {profile.allergens.map((a) => (
                    <li
                      key={a}
                      className="px-3 py-1 bg-brand-100 text-brand-500 rounded text-sm"
                    >
                      {a}
                    </li>
                  ))}
                </ul>
              )}
              <button
                onClick={() => setEditing(true)}
                className="text-brand-500 text-sm font-semibold hover:underline mt-2"
              >
                Düzenle
              </button>
            </>
          )}
        </div>
      </div>

      {error && (
        <div
          role="alert"
          className="mt-4 p-4 bg-brand-100 border border-brand-300 text-brand-500 rounded-lg"
        >
          {error}
        </div>
      )}

      <div className="flex flex-wrap gap-3 mt-6">
        <Link
          href="/upload"
          className="px-5 py-2 bg-brand-500 text-white font-semibold rounded-lg hover:bg-brand-600 transition"
        >
          Ürün analiz et
        </Link>
        <button
          onClick={async () => {
            await logout()
            onChange()
          }}
          className="px-5 py-2 border border-brand-200 rounded-lg hover:bg-brand-50 transition"
        >
          Çıkış yap
        </button>
      </div>

      {/* Veri hakları — gizlenmiş değil, görünür olmalı */}
      <section className="mt-10 pt-6 border-t border-brand-100">
        <h2 className="font-semibold mb-3">Verim üzerindeki haklarım</h2>
        <div className="flex flex-wrap gap-3">
          <a
            href={`${process.env.NEXT_PUBLIC_API_URL ?? ''}/profiles/me/export`}
            className="px-4 py-2 border border-brand-200 rounded-lg text-sm hover:bg-brand-50"
          >
            Verilerimi indir
          </a>
          {hasHealthData && (
            <button
              onClick={withdraw}
              disabled={busy}
              className="px-4 py-2 border border-clay/50 text-cocoa rounded-lg text-sm hover:bg-clay/10 disabled:opacity-50"
            >
              Alerjen verimi sil
            </button>
          )}
          <button
            onClick={removeAccount}
            disabled={busy}
            className="px-4 py-2 border border-brand-300 text-brand-500 rounded-lg text-sm hover:bg-brand-100 disabled:opacity-50"
          >
            Hesabımı sil
          </button>
        </div>
      </section>
    </div>
  )
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-sm text-ink-muted">{label}</p>
      <p className="font-semibold">{value}</p>
    </div>
  )
}
