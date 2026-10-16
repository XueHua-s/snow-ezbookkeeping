"""Synthetic environment and real child timeout regressions; no database access."""
import os
import pathlib
import subprocess
import sys
import tempfile
import time
import unittest
from assistant_migration_runtime import bounded_run, isolated_environment


class MigrationRuntimeTest(unittest.TestCase):
    def test_only_fixture_override_reaches_real_child(self):
        env = isolated_environment({'PATH': os.environ['PATH'],
            'EBK_DATABASE_HOST': 'unreachable-synthetic-host',
            'EBKCFP_DATABASE_PASSWD': '/synthetic/never-read',
            'EBK_WORK_DIR': '/wrong', 'NORMAL': 'retained'}, '/fixture')
        result = bounded_run([sys.executable, '-c',
            'import os,json; print(json.dumps(dict(os.environ)))'], env=env)
        import json
        actual = json.loads(result.stdout)
        self.assertEqual({k: v for k, v in actual.items()
            if k.startswith(('EBK_', 'EBKCFP_'))}, {'EBK_WORK_DIR': '/fixture'})
        self.assertEqual(actual['NORMAL'], 'retained')
        self.assertEqual(result.returncode, 0)

    def test_timeout_terminates_real_process_group(self):
        with tempfile.TemporaryDirectory() as directory:
            marker = pathlib.Path(directory) / 'pids'
            code = ('import os,time,pathlib,subprocess,sys; '
                    'child=subprocess.Popen([sys.executable,"-c","import time; time.sleep(30)"]); '
                    'pathlib.Path(sys.argv[1]).write_text(str(os.getpid())+" "+str(child.pid)); '
                    'time.sleep(30)')
            try:
                with self.assertRaises(subprocess.TimeoutExpired):
                    bounded_run([sys.executable, '-c', code, str(marker)], timeout=0.5)
                pids = [int(value) for value in marker.read_text().split()]
                deadline = time.monotonic() + 3
                while time.monotonic() < deadline:
                    states = [subprocess.run(['ps', '-o', 'stat=', '-p', str(pid)],
                              capture_output=True, text=True, timeout=1).stdout.strip()
                              for pid in pids]
                    if all(not state or state.startswith('Z') for state in states):
                        break
                    time.sleep(0.02)
                self.assertTrue(all(not state or state.startswith('Z') for state in states), states)
            finally:
                if marker.exists():
                    for value in marker.read_text().split():
                        try:
                            os.kill(int(value), 9)
                        except ProcessLookupError:
                            pass


if __name__ == '__main__':
    unittest.main()
