import { useEffect, useState } from "react"
import { SourcePicker } from "./views/sourcePicker"
import { TargetPicker } from "./views/targetPicker"
import { Rules } from "./views/rules"
import { Excute } from "./views/excute"
import { Monitor } from "./views/monitor"
import { Separator } from "@/components/ui/separator"
import { GetSettings, RunScan, ReloadSettings, ListDuplicateGroups, ExecuteTidy } from "../wailsjs/go/main/App"
import type { config, media, storage } from "../wailsjs/go/models"
import { EventsOff, EventsOn } from "../wailsjs/runtime/runtime"

function App() {
  const [settings, setSettings] = useState<config.Settings | null>(null)
  const [scanSummary, setScanSummary] = useState<media.Summary | null>(null)
  const [scanProgress, setScanProgress] = useState<any>(null)
  const [tidySummary, setTidySummary] = useState<media.TidySummary | null>(null)
  const [tidyProgress, setTidyProgress] = useState<any>(null)
  const [duplicates, setDuplicates] = useState<storage.DuplicateGroup[]>([])
  const [loadingScan, setLoadingScan] = useState(false)
  const [loadingTidy, setLoadingTidy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    refreshSettings()

    const offScan = EventsOn("scan:progress", (payload) => {
      setScanProgress(payload)
    })
    const offTidy = EventsOn("tidy:progress", (payload) => {
      setTidyProgress(payload)
    })

    return () => {
      EventsOff("scan:progress")
      EventsOff("tidy:progress")
      if (typeof offScan === "function") offScan()
      if (typeof offTidy === "function") offTidy()
    }
  }, [])

  const refreshSettings = async () => {
    try {
      setError(null)
      const loaded = await GetSettings()
      setSettings(loaded)
    } catch (err) {
      setError(String(err))
    }
  }

  const handleReload = async () => {
    try {
      setError(null)
      const reloaded = await ReloadSettings()
      setSettings(reloaded)
    } catch (err) {
      setError(String(err))
    }
  }

  const handleScan = async () => {
    setLoadingScan(true)
    setError(null)
    try {
      const summary = await RunScan()
      setScanSummary(summary)
      const groups = await ListDuplicateGroups()
      setDuplicates(groups)
    } catch (err) {
      setError(String(err))
    } finally {
      setLoadingScan(false)
    }
  }

  const handleDryRunTidy = async () => {
    if (duplicates.length === 0) {
      setError("No duplicates to tidy yet")
      return
    }

    const requests: media.MoveRequest[] = duplicates
      .flatMap((group) => group.Files.slice(1).map((file) => ({ mediaId: file.ID })))
      .filter(Boolean)

    if (requests.length === 0) {
      setError("Nothing to tidy — duplicates groups have single entries")
      return
    }

    setError(null)
    setLoadingTidy(true)
    try {
      const summary = await ExecuteTidy(requests, true)
      setTidySummary(summary)
    } catch (err) {
      setError(String(err))
    } finally {
      setLoadingTidy(false)
    }
  }

  const sources = settings?.Scan?.SourceFolders?.length
    ? settings.Scan.SourceFolders
    : settings?.History?.LastSourceFolder ?? []

  const extensions = settings?.Scan?.IncludeExtensions ?? []

  return (
    <main className='h-svh w-full overflow-hidden text-slate-100 bg-slate-950'>
      <h1 className='text-2xl m-2 font-bold text-right cursor-move'>photoTidyGo</h1>
      <div className='w-full h-[550px] flex gap-3 px-3 pb-3'>
        <div className='w-1/2 flex flex-col p-2 gap-3 justify-stretch'>
          <SourcePicker sources={sources} />
          <Separator className='m-1' />
          <TargetPicker targetPath={settings?.Target?.BaseFolder ?? ""} pattern={settings?.Target?.Pattern ?? ""} />
          <Separator className='m-1' />
          <Rules extensions={extensions} />
          <Separator className='m-1' />
          <Excute
            onScan={handleScan}
            onDryRun={handleDryRunTidy}
            onReload={handleReload}
            scanning={loadingScan}
            tidyRunning={loadingTidy}
            hasDuplicates={duplicates.length > 0}
          />
        </div>
        <div className='w-1/2 flex flex-col items-center justify-center'>
          <Monitor
            scanSummary={scanSummary}
            scanProgress={scanProgress}
            tidySummary={tidySummary}
            tidyProgress={tidyProgress}
            error={error}
          />
        </div>
      </div>
    </main>
  )
}

export default App
