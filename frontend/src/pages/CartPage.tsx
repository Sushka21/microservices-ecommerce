import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { cartApi } from '../lib/api'

type CartItem = { sku: number; count: number; name: string; price: number }

export default function CartPage({ userId }: { userId: number }) {
  const navigate = useNavigate()
  const [items, setItems] = useState<CartItem[]>([])
  const [totalPrice, setTotalPrice] = useState(0)
  const [loading, setLoading] = useState(true)
  const [checkingOut, setCheckingOut] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [deleting, setDeleting] = useState<number | null>(null)

  const loadCart = async () => {
    setLoading(true)
    setError(null)
    try {
      const res = await cartApi.listCart(userId)
      const list = res?.items ?? []
      setItems(list)
      setTotalPrice(res?.total_price ?? 0)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка загрузки корзины')
      setItems([])
      setTotalPrice(0)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCart()
  }, [userId])

  const removeItem = async (sku: number) => {
    setDeleting(sku)
    setError(null)
    try {
      await cartApi.deleteItem(userId, sku)
      await loadCart()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка удаления')
    } finally {
      setDeleting(null)
    }
  }

  const checkout = async () => {
    if (items.length === 0) return
    setCheckingOut(true)
    setError(null)
    try {
      const { order_id } = await cartApi.checkout(userId)
      navigate(`/order/${order_id}`)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка оформления заказа')
    } finally {
      setCheckingOut(false)
    }
  }

  if (loading) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
        <div className="h-8 w-48 bg-brand-surface rounded animate-pulse" />
        <div className="space-y-2">
          {[1, 2, 3].map((i) => (
            <div key={i} className="h-20 bg-brand-surface rounded-lg animate-pulse" />
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
      <h1 className="text-2xl font-semibold text-white mb-2">Корзина</h1>
      <p className="text-gray-400 mb-6">Товары текущего пользователя (User ID: {userId})</p>

      {error && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm"
        >
          {error}
        </motion.div>
      )}

      {items.length === 0 ? (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="text-gray-500 py-12 text-center"
        >
          Корзина пуста. Добавьте товары из каталога.
        </motion.p>
      ) : (
        <>
          <ul className="space-y-3 mb-6">
            {items.map((item, i) => (
              <motion.li
                key={`${item.sku}-${i}`}
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.04 }}
                className="flex items-center justify-between p-4 rounded-xl bg-brand-surface border border-white/5"
              >
                <div>
                  <span className="font-medium text-white">{item.name}</span>
                  <span className="text-gray-400 ml-2">SKU {item.sku} × {item.count}</span>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-brand-orange font-semibold">{item.price * item.count} ₽</span>
                  <button
                    type="button"
                    disabled={deleting === item.sku}
                    onClick={() => removeItem(item.sku)}
                    className="text-red-400 hover:text-red-300 text-sm disabled:opacity-50"
                  >
                    {deleting === item.sku ? '...' : 'Удалить'}
                  </button>
                </div>
              </motion.li>
            ))}
          </ul>

          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            className="flex flex-col sm:flex-row items-center justify-between gap-4 p-4 rounded-xl bg-brand-surface border border-brand-blue/20"
          >
            <span className="text-lg font-semibold text-white">Итого: {totalPrice} ₽</span>
            <button
              type="button"
              disabled={checkingOut}
              onClick={checkout}
              className="px-6 py-3 rounded-xl bg-brand-orange text-white font-semibold hover:opacity-90 disabled:opacity-60 transition-all shadow-lg shadow-brand-orange/20"
            >
              {checkingOut ? 'Оформление...' : 'Перейти к оплате'}
            </button>
          </motion.div>
        </>
      )}
    </motion.div>
  )
}
