#!/usr/bin/env node

const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

function runAnalyzer(options = {}) {
  const binPath = path.join(__dirname, 'bin', 'fsd-analyzer');
  
  if (!fs.existsSync(binPath)) {
    throw new Error('Ошибка: исполняемый файл не найден');
  }
  
  let args = '';
  
  if (options.config) {
    args += ` --config ${options.config}`;
  }
  
  try {
    return execSync(`${binPath}${args}`, { 
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