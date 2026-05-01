import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { orderApi, productApi } from '../lib/api'

const ORDER_STATUS: Record<number, string> = {
  0: 'Не указан',
  1: 'Новый',
  2: 'Ожидает оплаты',
  3: 'Ошибка',
  4: 'Оплачен',
  5: 'Отменён',
}

export default function OrderPage(_props: { userId: number }) {
  const { orderId } = useParams<{ orderId: string }>()
  const navigate = useNavigate()
  const id = orderId ? parseInt(orderId, 10) : NaN
  const [order, setOrder] = useState<{
    status: number
    user_id: number
    items: Array<{ sku: number; count: number }>
    created_at?: string
    updated_at?: string
  } | null>(null)
  const [loading, setLoading] = useState(true)
  const [total, setTotal] = useState<number | null>(null)
  const [action, setAction] = useState<'pay' | 'cancel' | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (Number.isNaN(id) || id < 1) {
      setLoading(false)
      return
    }
    let cancelled = false
    orderApi
      .get(id)
      .then((data) => {
        if (!cancelled) setOrder(data)
      })
      .catch((e) => {
        if (!cancelled) setError(e instanceof Error ? e.message : 'Ошибка загрузки заказа')
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })
    return () => {
      cancelled = true
    }
  }, [id])

  useEffect(() => {
    if (!order?.items?.length) {
      setTotal(0)
      return
    }
    let cancelled = false
    Promise.all(
      order.items.map((item) =>
        productApi.getProduct(item.sku).then((p) => p.price * item.count)
      )
    )
      .then((amounts) => {
        if (!cancelled) setTotal(amounts.reduce((a, b) => a + b, 0))
      })
      .catch(() => {
        if (!cancelled) setTotal(null)
      })
    return () => {
      cancelled = true
    }
  }, [order?.items])

  const doPay = async () => {
    if (Number.isNaN(id)) return
    setAction('pay')
    setError(null)
    try {
      await orderApi.pay(id)
      const data = await orderApi.get(id)
      setOrder(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка оплаты')
    } finally {
      setAction(null)
    }
  }

  const doCancel = async () => {
    if (Number.isNaN(id)) return
    setAction('cancel')
    setError(null)
    try {
      await orderApi.cancel(id)
      const data = await orderApi.get(id)
      setOrder(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Ошибка отмены')
    } finally {
      setAction(null)
    }
  }

  const canPay = order?.status === 2

  if (Number.isNaN(id)) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
        <p className="text-gray-500">Неверный заказ.</p>
        <button
          type="button"
          onClick={() => navigate('/')}
          className="mt-4 text-brand-blue hover:underline"
        >
          На главную
        </button>
      </motion.div>
    )
  }

  if (loading) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-4">
        <div className="h-8 w-64 bg-brand-surface rounded animate-pulse" />
        <div className="h-32 bg-brand-surface rounded-xl animate-pulse" />
      </motion.div>
    )
  }

  if (!order) {
    return (
      <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
        <p className="text-gray-500">Заказ не найден.</p>
        <button
          type="button"
          onClick={() => navigate('/cart')}
          className="mt-4 text-brand-blue hover:underline"
        >
          В корзину
        </button>
      </motion.div>
    )
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <h1 className="text-2xl font-semibold text-white mb-2">Заказ #{id}</h1>
      <p className="text-gray-400 mb-6">
        User ID: {order.user_id != null && order.user_id !== 0 ? order.user_id : '—'}
      </p>

      {error && (
        <motion.div
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-sm"
        >
          {error}
        </motion.div>
      )}

      <div className="rounded-xl bg-brand-surface border border-white/5 overflow-hidden mb-6">
        <div className="p-4 border-b border-white/5 flex items-center justify-between">
          <span className="text-gray-400">Статус</span>
          <span
            className={`font-semibold ${
              order.status === 4 ? 'text-green-400' : order.status === 5 ? 'text-gray-500' : 'text-brand-orange'
            }`}
          >
            {ORDER_STATUS[order.status] ?? order.status}
          </span>
        </div>
        <ul className="divide-y divide-white/5">
          {order.items.map((item, i) => (
            <motion.li
              key={i}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              transition={{ delay: i * 0.05 }}
              className="p-4 flex justify-between"
            >
              <span className="text-white">SKU {item.sku}</span>
              <span className="text-gray-400">× {item.count}</span>
            </motion.li>
          ))}
        </ul>
        <div className="p-4 border-t border-white/5 flex items-center justify-between">
          <span className="text-gray-400">Сумма</span>
          <span className="text-white font-semibold">
            {total != null ? `${total.toLocaleString('ru-RU')} ₽` : '—'}
          </span>
        </div>
      </div>

      {canPay && (
        <motion.div
          initial={{ opacity: 0, scale: 0.98 }}
          animate={{ opacity: 1, scale: 1 }}
          className="flex flex-wrap gap-3"
        >
          <button
            type="button"
            disabled={action !== null}
            onClick={doPay}
            className="px-6 py-3 rounded-xl bg-green-600 text-white font-semibold hover:bg-green-500 disabled:opacity-50 transition-all"
          >
            {action === 'pay' ? 'Оплата...' : 'Оплатить заказ'}
          </button>
          <button
            type="button"
            disabled={action !== null}
            onClick={doCancel}
            className="px-6 py-3 rounded-xl bg-brand-surface border border-red-500/50 text-red-400 font-semibold hover:bg-red-500/10 disabled:opacity-50 transition-all"
          >
            {action === 'cancel' ? 'Отмена...' : 'Отменить заказ'}
          </button>
        </motion.div>
      )}

      {order.status === 4 && (
        <motion.p
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className="text-green-400 font-medium"
        >
          Заказ оплачен.
        </motion.p>
      )}
      {order.status === 5 && (
        <motion.p initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-gray-500">
          Заказ отменён.
        </motion.p>
      )}

      <div className="mt-8 flex gap-4">
        <button
          type="button"
          onClick={() => navigate('/')}
          className="text-brand-blue hover:underline"
        >
          Каталог
        </button>
        <button
          type="button"
          onClick={() => navigate('/cart')}
          className="text-brand-blue hover:underline"
        >
          Корзина
        </button>
      </div>
    </motion.div>
  )
}
