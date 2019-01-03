const {gitDescribeSync} = require('git-describe');
const {version} = require('../../package.json');
const {resolve, relative} = require('path');
const {writeFileSync} = require('fs-extra');

const gitInfo = gitDescribeSync({
  dirtyMark: false,
  dirtySemver: false,
  longSemver: true,
});

gitInfo.packageVersion = version;
Object.assign(gitInfo.semver, {
  loose: false,
  options: {
    includePrerelease: false,
    loose: false,
  }
});

const file =
    resolve(__dirname, '..', '..', 'src/', 'app', 'frontend', 'environments', 'version.ts');
writeFileSync(
    file, `// Copyright 2017 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Please note that this file is autogenerated by "npm run postversion" script.

// clang-format off

import {VersionInfo} from '@api/frontendapi';

export const version: VersionInfo = ${JSON.stringify(gitInfo, null, 2).replace(/\"/g, '\'')};
`,
    {encoding: 'utf-8'});

console.log(`Version ${gitInfo.raw} saved to ${relative(resolve(__dirname, '..'), file)}`);
