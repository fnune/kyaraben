// Type check: verify that DAEMON_CHANNELS from electron/channels.ts
// contains only valid CommandType values from the generated daemon types.
// This file is not imported anywhere - it exists purely for compile-time checking.

import type { InvokeChannel } from '../../electron/channels'
import type { CommandType } from '../types/daemon.gen'

// Extract just the daemon channels from InvokeChannel
type DaemonChannel = Exclude<
  InvokeChannel,
  | 'get_install_status'
  | 'install_app'
  | 'open_path'
  | 'path_exists'
  | 'read_file'
  | 'get_bug_report_info'
  | 'launch_emulator'
  | 'open_log_tail'
  | 'launch_cli_uninstall'
  | 'check_for_updates'
  | 'download_update'
  | 'apply_update'
  | 'select_directory'
>

// This line will fail to compile if any DaemonChannel is not a valid CommandType
type _AssertDaemonChannelsAreValid = DaemonChannel extends CommandType ? true : never
const _check: _AssertDaemonChannelsAreValid = true
void _check
