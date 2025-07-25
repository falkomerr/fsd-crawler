#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');
const os = require('os');

function getExecutableName() {
  const platform = os.platform();
  const arch = os.arch();
  
  const normalizedPlatform = platform === 'win32' ? 'windows' : platform;
  
  let normalizedArch;
  switch (arch) {
    case 'x64':
      normalizedArch = 'amd64';
      break;
    case 'arm64':
      normalizedArch = 'arm64';
      break;
    case 'ia32':
      normalizedArch = 'amd64'; 
      break;
    default:
      normalizedArch = 'amd64'; 
  }
  
  const extension = normalizedPlatform === 'windows' ? '.exe' : '';
  return `fsd-analyzer-${normalizedPlatform}-${normalizedArch}${extension}`;
}

function runAnalyzer(options = {}) {
  const executableName = getExecutableName();
  const binPath = path.join(__dirname, 'bin', executableName);
  
  const fallbackPath = path.join(__dirname, 'bin', 'fsd-analyzer');
  
  let finalPath;
  if (fs.existsSync(binPath)) {
    finalPath = binPath;
  } else if (fs.existsSync(fallbackPath)) {
    finalPath = fallbackPath;
    console.warn(`⚠️  Используется исполняемый файл для неопределенной платформы. Для лучшей совместимости запустите 'npm run build:current'`);
  } else {
    throw new Error(`Ошибка: исполняемый файл не найден ни по пути ${binPath}, ни по пути ${fallbackPath}. Попробуйте запустить 'npm run build:current' или 'npm install'`);
  }
  
  let args = '';
  
  if (options.config) {
    args += ` --config ${options.config}`;
  }
  
  try {
    return execSync(`"${finalPath}"${args}`, { 
      stdio: options.silent ? 'ignore' : 'inherit',
      cwd: process.cwd()
    });
  } catch (error) {
    throw new Error(`Ошибка при запуске анализатора: ${error.message}`);
  }
}

if (require.main === module) {
  runAnalyzer();
} else {
  module.exports = runAnalyzer;
} 