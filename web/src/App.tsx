import React, { useState, useEffect, useCallback } from 'react'

interface Task {
  name: string
  display: string
  category: string
  status: string
  message?: string
  duration?: number
}

interface Stats {
  gold: number
  copper: number
  stamina: number
  max_stamina: number
  tasks_done: number
  tasks_total: number
}

interface AppState {
  running: boolean
  device_name: string
  device_ready: boolean
  uptime: number
}

export default function App() {
  const [state, setState] = useState<AppState>({
    running: false,
    device_name: '',
    device_ready: false,
    uptime: 0,
  })
  const [tasks, setTasks] = useState<Task[]>([])
  const [stats, setStats] = useState<Stats>({
    gold: 0,
    copper: 0,
    stamina: 0,
    max_stamina: 180,
    tasks_done: 0,
    tasks_total: 0,
  })
  const [loading, setLoading] = useState(false)

  // 获取任务列表
  const fetchTasks = useCallback(async () => {
    try {
      const res = await fetch('/api/tasks')
      const data = await res.json()
      setTasks(data)
    } catch (e) {
      console.error('Failed to fetch tasks:', e)
    }
  }, [])

  // 获取状态
  useEffect(() => {
    const fetchState = async () => {
      try {
        const res = await fetch('/api/status')
        const data = await res.json()
        setState(data)
      } catch (e) {
        console.error('Failed to fetch state:', e)
      }
    }

    const fetchStats = async () => {
      try {
        const res = await fetch('/api/stats')
        const data = await res.json()
        setStats(data)
      } catch (e) {
        console.error('Failed to fetch stats:', e)
      }
    }

    fetchState()
    fetchStats()
    fetchTasks()

    const interval = setInterval(fetchState, 2000)
    return () => clearInterval(interval)
  }, [fetchTasks])

  // 启动任务
  const handleStart = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/start', { method: 'POST' })
      const data = await res.json()
      if (res.ok) {
        setState(prev => ({ ...prev, running: true }))
      } else {
        alert(data.message || data.error)
      }
    } catch (e) {
      alert('启动失败')
    }
    setLoading(false)
  }

  // 停止任务
  const handleStop = async () => {
    setLoading(true)
    try {
      await fetch('/api/stop', { method: 'POST' })
      setState(prev => ({ ...prev, running: false }))
    } catch (e) {
      alert('停止失败')
    }
    setLoading(false)
  }

  // 连接设备
  const handleConnect = async () => {
    setLoading(true)
    try {
      const res = await fetch('/api/device/connect', { method: 'POST' })
      const data = await res.json()
      if (res.ok) {
        setState(prev => ({ ...prev, device_ready: true, device_name: data.device_name }))
      } else {
        alert(data.message || data.error)
      }
    } catch (e) {
      alert('连接失败')
    }
    setLoading(false)
  }

  // 执行单个任务
  const runTask = async (taskName: string) => {
    try {
      await fetch(`/api/task/${taskName}/run`, { method: 'POST' })
      fetchTasks()
    } catch (e) {
      alert('执行失败')
    }
  }

  const dailyTasks = tasks.filter(t => t.category === 'daily')
  const harvestTasks = tasks.filter(t => t.category === 'harvest')
  const weeklyTasks = tasks.filter(t => t.category === 'weekly')
  const eventTasks = tasks.filter(t => t.category === 'event')

  const formatUptime = (seconds: number) => {
    const h = Math.floor(seconds / 3600)
    const m = Math.floor((seconds % 3600) / 60)
    const s = Math.floor(seconds % 60)
    return `${h}h ${m}m ${s}s`
  }

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
    <div className="min-h-screen bg-gradient-to-br from-[#1a1a2e] to-[#16213e]">
      {/* Header */}
      <header className="border-b border-gray-700 bg-[#1a1a2e]/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <span className="text-3xl">🍥</span>
            <div>
              <h1 className="text-xl font-bold text-white">NarutoScript Next</h1>
              <div className="text-xs text-gray-400">运行时间: {formatUptime(state.uptime)}</div>
            </div>
          </div>
          <div className="flex items-center gap-4">
            {/* 设备状态 */}
            <div className="flex items-center gap-2">
              <span className={`w-2 h-2 rounded-full ${state.device_ready ? 'bg-green-500' : 'bg-red-500'}`}></span>
              <span className="text-sm text-gray-400">
                {state.device_ready ? state.device_name : '设备未连接'}
              </span>
            </div>
            {/* 运行状态 */}
            <span className={`px-3 py-1 rounded-full text-sm ${
              state.running ? 'bg-green-500/20 text-green-400' : 'bg-gray-500/20 text-gray-400'
            }`}>
              {state.running ? '运行中' : '已停止'}
            </span>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 space-y-6">
        {/* Stats */}
        <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <StatCard label="💰 金币" value={stats.gold.toLocaleString()} />
          <StatCard label="🪙 铜币" value={stats.copper.toLocaleString()} color="orange" />
          <StatCard label="⚡ 体力" value={`${stats.stamina}/${stats.max_stamina}`} color="blue" />
          <StatCard label="📋 任务" value={`${stats.tasks_done}/${stats.tasks_total}`} color="green" />
        </section>

        {/* Actions */}
        <section className="flex gap-4 flex-wrap">
          {!state.device_ready && (
            <button
              onClick={handleConnect}
              disabled={loading}
              className="bg-blue-600 hover:bg-blue-700 text-white py-3 px-6 rounded-xl font-medium transition disabled:opacity-50"
            >
              📱 连接设备
            </button>
          )}
          {state.running ? (
            <button
              onClick={handleStop}
              disabled={loading}
              className="flex-1 bg-red-600 hover:bg-red-700 text-white py-3 px-6 rounded-xl font-medium transition disabled:opacity-50"
            >
              ⏹️ 停止任务
            </button>
          ) : (
            <button
              onClick={handleStart}
              disabled={loading || !state.device_ready}
              className="flex-1 bg-orange-600 hover:bg-orange-700 text-white py-3 px-6 rounded-xl font-medium transition disabled:opacity-50"
            >
              ▶️ 执行全部
            </button>
          )}
        </section>

        {/* Tasks */}
        <section className="grid md:grid-cols-2 gap-6">
          <TaskGroup title="📅 日常任务" tasks={dailyTasks} statusColors={statusColors} statusIcons={statusIcons} onRun={runTask} disabled={!state.device_ready || state.running} />
          <TaskGroup title="🎁 收获系统" tasks={harvestTasks} statusColors={statusColors} statusIcons={statusIcons} onRun={runTask} disabled={!state.device_ready || state.running} />
          <TaskGroup title="📆 周常任务" tasks={weeklyTasks} statusColors={statusColors} statusIcons={statusIcons} onRun={runTask} disabled={!state.device_ready || state.running} />
          <TaskGroup title="🎉 限时活动" tasks={eventTasks} statusColors={statusColors} statusIcons={statusIcons} onRun={runTask} disabled={!state.device_ready || state.running} />
        </section>
      </main>

      <footer className="border-t border-gray-700 mt-8 py-4 text-center text-gray-500 text-sm">
        NarutoScript Next v1.0.0 | Go + React
      </footer>
    </div>
  )
}

function StatCard({ label, value, color = 'yellow' }: { label: string; value: string; color?: string }) {
  const colors: Record<string, string> = {
    yellow: 'text-yellow-400',
    orange: 'text-orange-400',
    blue: 'text-blue-400',
    green: 'text-green-400',
  }

  return (
    <div className="bg-[#252545] rounded-xl p-4 border border-gray-700 hover:border-gray-600 transition">
      <div className="text-gray-400 text-sm">{label}</div>
      <div className={`text-2xl font-bold ${colors[color]}`}>{value}</div>
    </div>
  )
}

function TaskGroup({
  title,
  tasks,
  statusColors,
  statusIcons,
  onRun,
  disabled,
}: {
  title: string
  tasks: Task[]
  statusColors: Record<string, string>
  statusIcons: Record<string, string>
  onRun: (name: string) => void
  disabled: boolean
}) {
  return (
    <div className="bg-[#252545] rounded-xl p-4 border border-gray-700">
      <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
        {title}
        <span className="text-xs bg-blue-500/20 text-blue-400 px-2 py-0.5 rounded-full">
          {tasks.length}
        </span>
      </h2>
      <div className="space-y-2">
        {tasks.map(task => (
          <div
            key={task.name}
            onClick={() => !disabled && task.status === 'idle' && onRun(task.name)}
            className={`flex items-center justify-between p-3 bg-[#1a1a2e] rounded-lg transition ${
              !disabled && task.status === 'idle' ? 'hover:bg-[#2a2a4e] cursor-pointer' : 'opacity-75'
            }`}
          >
            <div className="flex flex-col">
              <span className="text-gray-300">{task.display}</span>
              {task.message && <span className="text-xs text-gray-500">{task.message}</span>}
            </div>
            <span className={`text-xs px-2 py-1 rounded-full ${statusColors[task.status]}`}>
              {statusIcons[task.status]} {task.status}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
}