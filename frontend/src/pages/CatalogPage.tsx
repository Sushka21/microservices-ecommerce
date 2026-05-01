import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { motion } from 'framer-motion'
import { cartApi, productApi } from '../lib/api'

type ProductInfo = { sku: number; name: string; price: number }

interface CatalogPageProps {
  userId: number
  catalogSkus?: number[]
  catalogReady?: boolean
  catalogError?: string | null
}

export default function CatalogPage({ userId, catalogSkus = [], catalogReady = false, catalogError = null }: CatalogPageProps) {
  const [products, setProducts] = useState<ProductInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [adding, setAdding] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [bySku, setBySku] = useState({ sku: '', count: '1' })
  const [addedFlash, setAddedFlash] = useState<number | null>(null)

  useEffect(() => {
    if (!catalogReady) return
    if (catalogSkus.length === 0) {
      setLoading(false)
      return
    }
    let cancelled = false
    setLoading(true)
    Promise.all(
      catalogSkus.map(async (sku) => {
        try {
          const p = await productApi.getProduct(sku)
          return { sku, name: p.name, price: p.price }
        } catch {
          return null
        }
      })
    )
      .then((list) => {
        if (!cancelled) setProducts(list.filter((p): p is ProductInfo => p !== null))
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [catalogReady, catalogSkus])

  const addToCart = async (sku: number, count: number) => {
    if (count < 1) return
    setAdding(sku)
    setError(null)
    try {
      await cartApi.addItem(userId, sku, count)
      setAddedFlash(sku)
      setTimeout(() => setAddedFlash(null), 600)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка добавления')
    } finally {
      setAdding(null)
    }
  }

  const addBySku = () => {
    const sku = parseInt(bySku.sku, 10)
    const count = parseInt(bySku.count, 10) || 1
    if (Number.isNaN(sku) || sku < 0) return
    addToCart(sku, count)
  }

  if (!catalogReady || loading) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
        <h1 className="text-2xl font-semibold text-white mb-2">Каталог</h1>
        <p className="text-gray-400 mb-6">
          {!catalogReady ? 'Создаём товары и остатки…' : 'Загрузка карточек…'}
        </p>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-56 rounded-2xl bg-brand-surface animate-pulse ring-1 ring-white/10" />
          ))}
        </div>
      </motion.div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      {catalogError && (
        <div className="mb-6 p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400">
          <p className="font-semibold">Не удалось инициализировать магазин</p>
          <p className="mt-1 text-sm">{catalogError}</p>
        </div>
      )}

      <div className="flex flex-wrap items-center justify-between gap-4 mb-6">
        <div>
          <h1 className="text-2xl font-semibold text-white mb-1">Каталог</h1>
          <p className="text-gray-400">Клик по карточке — +1 в корзину. Затем перейдите в корзину.</p>
        </div>
        <Link
          to="/cart"
          className="px-4 py-2.5 rounded-xl bg-brand-orange text-white font-medium hover:opacity-90 transition-opacity"
        >
          Перейти в корзину
        </Link>
      </div>

      {error && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm"
        >
          {error}
        </motion.div>
      )}

      <div className="mb-8 p-4 rounded-xl bg-brand-surface border border-white/5">
        <p className="text-sm text-gray-400 mb-3">Добавить по SKU</p>
        <div className="flex flex-wrap gap-2">
          <input
            type="number"
            min={0}
            placeholder="SKU"
            value={bySku.sku}
            onChange={(e) => setBySku((s) => ({ ...s, sku: e.target.value }))}
            className="w-24 px-3 py-2 rounded-lg bg-brand-dark border border-white/10 text-white placeholder-gray-500 focus:border-brand-blue outline-none"
          />
          <input
            type="number"
            min={1}
            placeholder="Кол-во"
            value={bySku.count}
            onChange={(e) => setBySku((s) => ({ ...s, count: e.target.value }))}
            className="w-20 px-3 py-2 rounded-lg bg-brand-dark border border-white/10 text-white placeholder-gray-500 focus:border-brand-blue outline-none"
          />
          <button
            type="button"
            onClick={addBySku}
            className="px-4 py-2 rounded-lg bg-brand-orange text-white font-medium hover:opacity-90 transition-opacity"
          >
            В корзину
          </button>
        </div>
      </div>

      {products.length === 0 && (
        <p className="text-gray-500 py-6">
          Каталог пуст. Убедитесь, что бэкенд (LOMS, cart) запущен, или добавьте товар по SKU ниже.
        </p>
      )}

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
        {products.map((p, i) => (
          <motion.article
            key={p.sku}
            role="button"
            tabIndex={0}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.08 }}
            onClick={() => addToCart(p.sku, 1)}
            onKeyDown={(e) => e.key === 'Enter' && addToCart(p.sku, 1)}
            className={`relative rounded-2xl p-6 transition-all duration-300 cursor-pointer select-none ${
              addedFlash === p.sku
                ? 'ring-2 ring-brand-orange shadow-lg shadow-brand-orange/20'
                : 'ring-1 ring-white/10 hover:ring-brand-blue/50 hover:shadow-xl hover:shadow-brand-blue/5'
            } bg-brand-surface`}
          >
            <div className="absolute inset-0 rounded-2xl border-2 border-brand-blue/20 pointer-events-none" aria-hidden />
            <div className="relative">
              <span className="text-xs font-medium text-brand-blue uppercase tracking-wider">SKU {p.sku}</span>
              <h3 className="mt-2 text-xl font-semibold text-white">{p.name}</h3>
              <p className="mt-2 text-2xl font-bold text-brand-orange">{p.price.toLocaleString('ru-RU')} ₽</p>
              <p className="mt-1 text-sm text-gray-500">Клик: +1 в корзину</p>
              <div className="mt-5 flex gap-2" onClick={(e) => e.stopPropagation()}>
                {[1, 2, 5].map((q) => (
                  <button
                    key={q}
                    type="button"
                    disabled={adding === p.sku}
                    onClick={() => addToCart(p.sku, q)}
                    className="flex-1 py-2.5 rounded-xl bg-brand-blue/20 text-brand-blue font-medium hover:bg-brand-blue/30 disabled:opacity-50 transition-all border border-brand-blue/30"
                  >
                    {adding === p.sku ? '...' : `+${q}`}
                  </button>
                ))}
              </div>
            </div>
          </motion.article>
        ))}
      </div>
    </motion.div>
  )
}
