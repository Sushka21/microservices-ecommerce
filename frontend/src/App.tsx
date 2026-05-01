import { useState, useEffect } from 'react'
import { Routes, Route, Link, useNavigate, useLocation } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { getUserId, resetUserId } from './lib/cookies'
import { productApi, stockApi } from './lib/api'
import CartPage from './pages/CartPage'
import CatalogPage from './pages/CatalogPage'
import OrderPage from './pages/OrderPage'

const pageTransition = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -4 },
  transition: { duration: 0.25 },
}

const SEED_PRODUCTS = [
  { name: 'Кроссовки', price: 49000, stock: 100 },
  { name: 'Майка', price: 27000, stock: 100 },
  { name: 'Стул', price: 11000, stock: 50 },
]

function App() {
  const [userId, setUserIdState] = useState(getUserId)
  const [catalogSkus, setCatalogSkus] = useState<number[]>([])
  const [catalogReady, setCatalogReady] = useState(false)
  const [catalogError, setCatalogError] = useState<string | null>(null)
  const navigate = useNavigate()
  const location = useLocation()

  useEffect(() => {
    const init = async () => {
      const skus: number[] = []
      try {
        for (const p of SEED_PRODUCTS) {
          const { sku } = await productApi.createProduct(p.name, p.price)
          if (sku >= 1) {
            await stockApi.setStock(sku, p.stock)
            skus.push(sku)
          }
        }
        if (skus.length === 0) {
          setCatalogError('Не удалось создать товары или проставить остатки. Проверьте, что бэкенд (LOMS) запущен.')
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Неизвестная ошибка'
        setCatalogError(`Не удалось инициализировать магазин: ${msg}`)
      } finally {
        setCatalogSkus(skus)
        setCatalogReady(true)
      }
    }
    init()
  }, [])

  const handleResetUser = () => {
    const newId = resetUserId()
    setUserIdState(newId)
    navigate('/')
  }

  return (
    <div className="min-h-screen flex flex-col bg-brand-dark">
      <header className="sticky top-0 z-50 border-b border-white/5 bg-brand-dark/95 backdrop-blur">
        <div className="max-w-6xl mx-auto px-4 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-2 text-xl font-semibold text-white">
            <span className="text-brand-blue">Igoroutine</span>
            <span className="text-brand-orange">Shop</span>
          </Link>
          <nav className="flex items-center gap-4">
            <Link
              to="/cart"
              className="px-3 py-2 rounded-lg text-brand-blue hover:bg-brand-blue/10 transition-colors font-medium"
            >
              Корзина
            </Link>
            <div className="flex items-center gap-2 pl-4 border-l border-white/10">
              <span className="text-gray-400 text-sm">User ID:</span>
              <span className="font-mono text-brand-blue bg-brand-surface px-2 py-1 rounded text-sm">
                {userId}
              </span>
              <button
                type="button"
                onClick={handleResetUser}
                className="text-xs px-2 py-1 rounded bg-brand-surface hover:bg-brand-orange/20 text-brand-orange border border-brand-orange/30 transition-colors"
              >
                Сбросить
              </button>
            </div>
          </nav>
        </div>
      </header>

      <main className="flex-1 max-w-6xl w-full mx-auto px-4 py-8">
        <AnimatePresence mode="wait">
          <Routes location={location} key={location.pathname}>
            <Route
              path="/"
              element={
                <motion.div {...pageTransition}>
                  <CatalogPage userId={userId} catalogSkus={catalogSkus} catalogReady={catalogReady} catalogError={catalogError} />
                </motion.div>
              }
            />
            <Route
              path="/cart"
              element={
                <motion.div {...pageTransition}>
                  <CartPage userId={userId} />
                </motion.div>
              }
            />
            <Route
              path="/order/:orderId"
              element={
                <motion.div {...pageTransition}>
                  <OrderPage userId={userId} />
                </motion.div>
              }
            />
          </Routes>
        </AnimatePresence>
      </main>

      <footer className="border-t border-white/5 py-4 text-center text-gray-500 text-sm">
        E-Commerce · Без авторизации (user ID в cookie)
      </footer>
    </div>
  )
}

export default App
