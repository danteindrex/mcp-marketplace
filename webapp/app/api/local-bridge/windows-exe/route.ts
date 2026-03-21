import { promises as fs } from 'fs'
import os from 'os'
import path from 'path'
import { execFile } from 'child_process'
import { promisify } from 'util'
import { NextResponse } from 'next/server'

export const runtime = 'nodejs'

const execFileAsync = promisify(execFile)

async function ensureBridgeBinary() {
  const repoRoot = path.resolve(process.cwd(), '..')
  const backendRoot = path.join(repoRoot, 'backend')
  const outputDir = path.join(os.tmpdir(), 'mcp-marketplace')
  const outputPath = path.join(outputDir, 'mcp-local-bridge.exe')

  await fs.mkdir(outputDir, { recursive: true })

  try {
    await execFileAsync('go', ['build', '-o', outputPath, './cmd/local-bridge'], {
      cwd: backendRoot,
      windowsHide: true,
    })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'go build failed'
    throw new Error(`Unable to build MCP Local Bridge: ${message}`)
  }

  return outputPath
}

export async function GET() {
  try {
    const outputPath = await ensureBridgeBinary()
    const binary = await fs.readFile(outputPath)

    return new NextResponse(binary, {
      status: 200,
      headers: {
        'Content-Type': 'application/vnd.microsoft.portable-executable',
        'Content-Disposition': 'attachment; filename="mcp-local-bridge.exe"',
        'Cache-Control': 'no-store',
      },
    })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unable to build MCP Local Bridge'
    return NextResponse.json({ error: message }, { status: 500 })
  }
}
