"""Bounded subprocess boundary for the isolated migration verification CLI."""
import os
import signal
import subprocess


def isolated_environment(source, root):
    """INI fixtures own all app settings; inherited app overrides cannot escape."""
    env = {key: value for key, value in source.items()
           if not key.startswith(('EBK_', 'EBKCFP_'))}
    env.update(BUILD_PIPELINE='1', EBK_WORK_DIR=str(root))
    return env


def bounded_run(cmd, *, data=None, cwd=None, env=None, timeout=180):
    """Kill the task process group on timeout, including compiler descendants."""
    with subprocess.Popen(cmd, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                          stderr=subprocess.PIPE, cwd=cwd, env=env,
                          start_new_session=True) as process:
        try:
            stdout, stderr = process.communicate(data, timeout=timeout)
        except subprocess.TimeoutExpired:
            try:
                os.killpg(process.pid, signal.SIGKILL)
            except ProcessLookupError:
                pass
            process.communicate()
            raise
        return subprocess.CompletedProcess(cmd, process.returncode, stdout, stderr)
