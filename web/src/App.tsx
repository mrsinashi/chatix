import { motion } from 'motion/react'

export default function App() {
  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="text-center"
      >
        <h1 className="text-3xl font-semibold text-gray-900">Chatix</h1>
        <p className="mt-2 text-gray-500">Клиент загружен. Каркас готов.</p>
      </motion.div>
    </div>
  )
}