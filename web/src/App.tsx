import { useState, useEffect } from 'react'

interface Task {
  name: string
  display: string
  category: string
  status: string
}

interface Stats {
  gold: number
  copper: number
  stamina: number
  tasks_done: number
  tasks_total: number
}

function App() {
  const [running, setRunning] = useState(false)
  const [tasks, setTasks] = useState<Task[]>([])
  const [stats, setStats] = useState<Stats>({
    gold: 0,
    copper: 0,
    stamina: 0,
    tasks_done: 0,
    tasks_total: 0,
  })

  // 获取任务列表
  useEffect(() => {
    fetch('/api/tasks')
      .then(res => res.json())
      .then(data => setTasks(data))
  }, [])

  // 获取状态
  useEffect(() => {
    const interval = setInterval(() => {
      fetch('/api/status')
        .then(res => res.json())
        .then(data => setRunning(data.running))
      
      fetch('/api/stats')
        .then(res => res.json())
        .then(data => setStats(data))
    }, 1000)

    return () => clearInterval(interval)
  }, [])

  const handleStart = async () => {
    await fetch('/api/start', { method: 'POST' })
    setRunning(true)
  }

  const handleStop = async () => {
    await fetch('/api/stop', { method: 'POST' })
    setRunning(false)
  }

  const dailyTasks = tasks.filter(t => t.category === 'daily')
  const harvestTasks = tasks.filter(t => t.category === 'harvest')
  const weeklyTasks = tasks.filter(t => t.category === 'weekly')
  const eventTasks = tasks.filter(t => t.category === 'event')

  return (
    <div className="min-h-screen bg-gradient-to-br from-[#1a1a2e] to-[#16213e]">
      {/* Header */}
      <header className="border-b border-gray-700 bg-[#1a1a2e]/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-3xl">🍥</span>
            <h1 className="text-xl font-bold text-white">NarutoScript Next</h1>
          </div>
          <div className="flex items-center gap-4">
            <span className={`px-3 py-1 rounded-full text-sm ${running ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'}`}>
              {running ? '运行中' : '已停止'}
            </span>
            <button 
              onClick={() => window.open('https://github.com/xfengyin/narutoscript-next', '_blank')}
              className="p-2 hover:bg-gray-700 rounded-lg transition"
            >
              ⚙️
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 space-y-6">
        {/* Stats */}
        <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <div className="text-gray-400 text-sm">💰 金币</div>
            <div className="text-2xl font-bold text-yellow-400">{stats.gold.toLocaleString()}</div>
          </div>
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <div className="text-gray-400 text-sm">🪙 铜币</div>
            <div className="text-2xl font-bold text-orange-400">{stats.copper.toLocaleString()}</div>
          </div>
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <div className="text-gray-400 text-sm">⚡ 体力</div>
            <div className="text-2xl font-bold text-blue-400">{stats.stamina}/180</div>
          </div>
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <div className="text-gray-400 text-sm">📋 任务</div>
            <div className="text-2xl font-bold text-green-400">{stats.tasks_done}/{stats.tasks_total}</div>
          </div>
        </section>

        {/* Actions */}
        <section className="flex gap-4">
          {running ? (
            <button 
              onClick={handleStop}
              className="flex-1 bg-red-600 hover:bg-red-700 text-white py-3 px-6 rounded-xl font-medium transition flex items-center justify-center gap-2"
            >
              ⏹️ 停止任务
            </button>
          ) : (
            <button 
              onClick={handleStart}
              className="flex-1 bg-naruto-orange hover:bg-orange-600 text-white py-3 px-6 rounded-xl font-medium transition flex items-center justify-center gap-2"
            >
              ▶️ 执行全部
            </button>
          )}
        </section>

        {/* Tasks */}
        <section className="grid md:grid-cols-2 gap-6">
          {/* Daily */}
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              📅 日常任务
              <span className="text-xs bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full">{dailyTasks.length}</span>
            </h2>
            <div className="space-y-2">
              {dailyTasks.map(task => (
                <TaskItem key={task.name} task={task} />
              ))}
            </div>
          </div>

          {/* Harvest */}
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              🎁 收获系统
              <span className="text-xs bg-green-500/20 text-green-400 px-2 py-0.5 rounded-full">{harvestTasks.length}</span>
            </h2>
            <div className="space-y-2">
              {harvestTasks.map(task => (
                <TaskItem key={task.name} task={task} />
              ))}
            </div>
          </div>

          {/* Weekly */}
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              📆 周常任务
              <span className="text-xs bg-purple-500/20 text-purple-400 px-2 py-0.5 rounded-full">{weeklyTasks.length}</span>
            </h2>
            <div className="space-y-2">
              {weeklyTasks.map(task => (
                <TaskItem key={task.name} task={task} />
              ))}
            </div>
          </div>

          {/* Event */}
          <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
            <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
              🎉 限时活动
              <span className="text-xs bg-yellow-500/20 text-yellow-400 px-2 py-0.5 rounded-full">{eventTasks.length}</span>
            </h2>
            <div className="space-y-2">
              {eventTasks.map(task => (
                <TaskItem key={task.name} task={task} />
              ))}
            </div>
          </div>
        </section>
      </main>

      {/* Footer */}
      <footer className="border-t border-gray-700 mt-8 py-4 text-center text-gray-500 text-sm">
        NarutoScript Next v1.0.0 | 基于 Go + React 构建
      </footer>
    </div>
  )
}

function TaskItem({ task }: { task: Task }) {
  const statusColors: Record<string, string> = {
    idle: 'bg-gray-500/20 text-gray-400',
    running: 'bg-blue-500/20 text-blue-400',
    success: 'bg-green-500/20 text-green-400',
    failed: 'bg-red-500/20 text-red-400',
    waiting: 'bg-yellow-500/20 text-yellow-400',
  }

  const statusIcons: Record<string, string> = {
    idle: '⏸️',
    running: '⏳',
    success: '✅',
    failed: '❌',
    waiting: '⏰',
  }

  return (
    <div className="flex items-center justify-between p-3 bg-[#1a1a2e] rounded-lg hover:bg-[#2a2a4e] transition cursor-pointer">
      <span className="text-gray-300">{task.display}</span>
      <span className={`text-xs px-2 py-1 rounded-full ${statusColors[task.status]}`}>
        {statusIcons[task.status]} {task.status}
      </span>
    </div>
  )
}

export default App