// app/profile/page.tsx — cilt profili oluştur ve sakla
'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import {
  createProfile,
  getProfile,
  SKIN_TYPES,
  skinTypeLabel,
  ApiError,
  type UserProfile,
} from '@/lib/api'

// MVP'de oturum yok, bu yüzden profil kimliği localStorage'da tutulur.
// Faz 2'de yerini gerçek bir oturum alacak.
const STORAGE_KEY = 'beauty:profile_id'

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null)
  const [email, setEmail] = useState('')
  const [skinType, setSkinType] = useState<string>('sensitive')
  const [allergens, setAllergens] = useState('')
  const [loading, setLoading] = useState(false)
  const [restoring, setRestoring] = useState(true)
  const [error, setError] = useState('')

  // İlk render'da kayıtlı profili geri yükle.
  useEffect(() => {
    const saved = window.localStorage.getItem(STORAGE_KEY)
    if (!saved) {
      setRestoring(false)
      return
    }

    getProfile(saved)
      .then(setProfile)
      .catch(() => window.localStorage.removeItem(STORAGE_KEY))
      .finally(() => setRestoring(false))
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')

    try {
      const created = await createProfile({
        email: email.trim() || undefined,
        skin_type: skinType,
        allergens: allergens
          .split(',')
          .map((a) => a.trim())
          .filter(Boolean),
      })
      window.localStorage.setItem(STORAGE_KEY, String(created.id))
      setProfile(created)
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Profiliniz kaydedilemedi.')
    } finally {
      setLoading(false)
    }
  }

  const handleReset = () => {
    window.localStorage.removeItem(STORAGE_KEY)
    setProfile(null)
    setEmail('')
    setAllergens('')
  }

  if (restoring) {
    return <p className="py-12 text-center text-mauve">Profiliniz yükleniyor…</p>
  }

  if (profile) {
    return (
      <div className="max-w-2xl mx-auto py-12">
        <h1 className="text-4xl font-bold mb-2">Profilim</h1>
        <p className="text-espresso/70 mb-8">
          Bu cihazda kayıtlı. Yükleme ve ürün sayfaları, tepki verdiğiniz
          içerikleri işaretlemek için bunu kullanır.
        </p>

        <div className="border border-brand-100 rounded-xl p-6 bg-white space-y-4">
          <Field label="Profil kimliği" value={`#${profile.id}`} />
          {profile.email && <Field label="E-posta" value={profile.email} />}
          <Field
            label="Cilt tipi"
            value={
              profile.skin_type ? skinTypeLabel(profile.skin_type) : 'belirtilmedi'
            }
          />

          <div>
            <p className="text-sm text-mauve mb-2">Alerjenler</p>
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
          </div>
        </div>

        <div className="flex gap-3 mt-6">
          <Link
            href="/upload"
            className="px-5 py-2 bg-brand-400 text-white font-semibold rounded-lg hover:bg-brand-500 transition"
          >
            Ürün analiz et
          </Link>
          <button
            onClick={handleReset}
            className="px-5 py-2 border border-brand-200 rounded-lg hover:bg-brand-50 transition"
          >
            Baştan başla
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="max-w-2xl mx-auto py-12">
      <h1 className="text-4xl font-bold mb-2">Profilini Oluştur</h1>
      <p className="text-espresso/70 mb-8">
        Cildinin neye tepki verdiğini söyle; uygulamanın her yerinde o
        içerikleri senin için işaretleyelim.
      </p>

      <form
        onSubmit={handleSubmit}
        className="border border-brand-100 rounded-xl p-6 bg-white space-y-5"
      >
        <div>
          <label htmlFor="email" className="block text-sm font-semibold mb-1">
            E-posta <span className="font-normal text-mist">(isteğe bağlı)</span>
          </label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="sen@ornek.com"
            className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-400"
          />
        </div>

        <div>
          <label htmlFor="skin_type" className="block text-sm font-semibold mb-1">
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

        <div>
          <label htmlFor="allergens" className="block text-sm font-semibold mb-1">
            Bilinen alerjenlerin
          </label>
          <input
            id="allergens"
            value={allergens}
            onChange={(e) => setAllergens(e.target.value)}
            placeholder="parfüm, nikel, formaldehit"
            className="w-full px-4 py-2 border border-brand-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-brand-400"
          />
          <p className="text-sm text-mist mt-1">Virgülle ayırın.</p>
        </div>

        {error && (
          <div className="p-4 bg-brand-100 border border-brand-300 text-brand-500 rounded-lg">
            {error}
          </div>
        )}

        <button
          type="submit"
          disabled={loading}
          className="w-full py-3 bg-brand-400 text-white font-semibold rounded-lg hover:bg-brand-500 disabled:opacity-50 disabled:cursor-not-allowed transition"
        >
          {loading ? 'Kaydediliyor…' : 'Profili kaydet'}
        </button>
      </form>
    </div>
  )
}

function Field({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <p className="text-sm text-mauve">{label}</p>
      <p className="font-semibold">{value}</p>
    </div>
  )
}
