import importlib.util
import io
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
import urllib.error


SPEC = importlib.util.spec_from_file_location(
    "semgrep_baseline", Path(__file__).with_name("semgrep_baseline.py")
)
baseline = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(baseline)


def git(cwd, *args):
    return subprocess.check_output(
        ["git", "-C", str(cwd), *args], text=True, stderr=subprocess.PIPE
    ).strip()


class BaselineTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.root = Path(self.tmp.name)
        self.origin = self.root / "origin.git"
        self.seed = self.root / "seed"
        self.repo = self.root / "checkout"
        subprocess.run(["git", "init", "--quiet", "--bare", str(self.origin)], check=True)
        subprocess.run(["git", "init", "--quiet", str(self.seed)], check=True)
        git(self.seed, "config", "user.name", "CI Test")
        git(self.seed, "config", "user.email", "ci-test@example.invalid")

        def commit(name):
            (self.seed / name).write_text(name)
            git(self.seed, "add", name)
            git(self.seed, "commit", "--quiet", "-m", name)
            return git(self.seed, "rev-parse", "HEAD")

        self.original = commit("original")
        git(self.seed, "checkout", "--quiet", "-b", "feature")
        self.v7_fork = commit("first-change")
        self.head = commit("second-change")
        git(self.seed, "checkout", "--quiet", "-b", "celo-contracts/v6.0.0", self.original)
        self.v6_tip = commit("v6-advance")
        git(self.seed, "checkout", "--quiet", "-b", "celo-contracts/v7.0.0", self.v7_fork)
        self.v7_tip = commit("v7-advance")
        git(self.seed, "remote", "add", "origin", str(self.origin))
        git(self.seed, "push", "--quiet", "--all", "origin")
        subprocess.run(
            ["git", "clone", "--quiet", "--shared", "--branch", "feature", str(self.origin), str(self.repo)],
            check=True,
        )
        self.publish_merge(self.v6_tip)
        self.env_file = self.root / "bash-env"
        self.env_file.touch()
        self.env = {
            "PATH": os.environ["PATH"],
            "BASH_ENV": str(self.env_file),
            "CIRCLE_PROJECT_USERNAME": "celo-org",
            "CIRCLE_PROJECT_REPONAME": "optimism",
            "CIRCLE_PULL_REQUEST": "https://github.com/celo-org/optimism/pull/570",
            "SEMGREP_BASELINE_REF": "develop",
        }

    def publish_merge(self, base, head=None):
        git(self.seed, "checkout", "--quiet", "--detach", base)
        git(self.seed, "merge", "--quiet", "--no-ff", "--no-edit", head or self.head)
        self.merge = git(self.seed, "rev-parse", "HEAD")
        git(self.seed, "push", "--quiet", "--force", "origin", f"{self.merge}:refs/pull/570/merge")

    def configure(self):
        previous = Path.cwd()
        try:
            os.chdir(self.repo)
            with urllib.error.HTTPError("https://api.github.com", 403, "Forbidden", {}, io.BytesIO()) as denied:
                with patch.dict(os.environ, self.env, clear=True):
                    with patch("urllib.request.urlopen", side_effect=denied):
                        try:
                            baseline.configure_baseline()
                        except urllib.error.HTTPError:
                            self.fail("Baseline lookup must work when the GitHub REST API is forbidden")
        finally:
            os.chdir(previous)

    def effective_baseline(self):
        return subprocess.check_output(
            ["bash", "-c", '. "$BASH_ENV"; printf "%s" "${SEMGREP_BASELINE_REF-unset}"'],
            env=self.env,
            text=True,
        )

    def test_pr_base_resolves_when_github_rest_api_is_forbidden(self):
        self.configure()
        self.assertEqual(self.original, self.effective_baseline())

    def test_retargeted_version_uses_its_own_merge_base(self):
        self.publish_merge(self.v7_tip)
        self.configure()
        self.assertEqual(self.v7_fork, self.effective_baseline())

    def test_shallow_checkout_can_resolve_the_common_ancestor(self):
        shallow = self.root / "shallow"
        subprocess.run(
            ["git", "clone", "--quiet", "--depth=1", "--branch", "feature", self.origin.as_uri(), str(shallow)],
            check=True,
        )
        self.repo = shallow
        self.configure()
        self.assertEqual(self.original, self.effective_baseline())

    def test_non_pr_push_clears_an_inherited_baseline_for_full_scan(self):
        self.env.pop("CIRCLE_PULL_REQUEST")
        self.configure()
        self.assertEqual("unset", self.effective_baseline())

    def test_invalid_pr_number_fails_without_exporting_a_baseline(self):
        self.env["CIRCLE_PULL_REQUEST"] = "https://github.com/celo-org/optimism/pull/--bad"
        with self.assertRaises(ValueError):
            self.configure()
        self.assertEqual("", self.env_file.read_text())

    def test_missing_merge_ref_fails_without_exporting_a_baseline(self):
        git(self.seed, "push", "--quiet", "origin", ":refs/pull/570/merge")
        with self.assertRaises(subprocess.CalledProcessError):
            self.configure()
        self.assertEqual("", self.env_file.read_text())

    def test_non_merge_ref_fails_without_exporting_a_baseline(self):
        git(self.seed, "push", "--quiet", "--force", "origin", f"{self.head}:refs/pull/570/merge")
        with self.assertRaises(ValueError):
            self.configure()
        self.assertEqual("", self.env_file.read_text())

    def test_merge_for_a_different_pr_head_is_rejected(self):
        self.publish_merge(self.v6_tip, self.v7_fork)
        with self.assertRaises(ValueError):
            self.configure()
        self.assertEqual("", self.env_file.read_text())


if __name__ == "__main__":
    unittest.main()
