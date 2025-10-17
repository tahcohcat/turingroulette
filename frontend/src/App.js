import React, { useState, useEffect, useRef } from 'react';
import { Brain, Zap, AlertCircle, Trophy, XCircle, Settings, BarChart3, Award } from 'lucide-react';

export default function TuringRoulette() {
  const [gameState, setGameState] = useState('setup');
  const [riddle, setRiddle] = useState('');
  const [answer, setAnswer] = useState('');
  const [clues, setClues] = useState(['', '', '']);
  const [difficulty, setDifficulty] = useState('medium');
  const [modelOutputs, setModelOutputs] = useState({});
  const [modelResults, setModelResults] = useState({});
  const [modelHistory, setModelHistory] = useState({});
  const [currentRound, setCurrentRound] = useState(0);
  const [gameResult, setGameResult] = useState(null);
  const [gameMessage, setGameMessage] = useState('');
  const [config, setConfig] = useState(null);
  const [stats, setStats] = useState(null);
  const [leaderboard, setLeaderboard] = useState([]);
  const [loading, setLoading] = useState(true);
  const ws = useRef(null);

  useEffect(() => {
    fetchConfig();
    fetchStats();
    fetchLeaderboard();
    return () => {
      if (ws.current) {
        ws.current.close();
      }
    };
  }, []);

  const fetchConfig = async () => {
    try {
      const response = await fetch('http://localhost:8080/config');
      const data = await response.json();
      setConfig(data);
      
      const outputs = {};
      data.models.forEach(model => {
        outputs[model.name] = '';
      });
      setModelOutputs(outputs);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching config:', error);
      setLoading(false);
    }
  };

  const fetchStats = async () => {
    try {
      const response = await fetch('http://localhost:8080/stats');
      const data = await response.json();
      setStats(data);
    } catch (error) {
      console.error('Error fetching stats:', error);
    }
  };

  const fetchLeaderboard = async () => {
    try {
      const response = await fetch('http://localhost:8080/leaderboard');
      const data = await response.json();
      setLeaderboard(data || []);
    } catch (error) {
      console.error('Error fetching leaderboard:', error);
    }
  };

  const connectWebSocket = () => {
    ws.current = new WebSocket('ws://localhost:8080/ws');
    
    ws.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'guess') {
        setModelOutputs(prev => ({
          ...prev,
          [data.model]: prev[data.model] + data.content
        }));
      } else if (data.type === 'result') {
        setModelResults(prev => ({
          ...prev,
          [data.model]: data.content === 'true'
        }));
      } else if (data.type === 'gameFinished') {
        if (data.modelStates) {
          setModelHistory(data.modelStates);
        }
        setGameResult({
          playerWins: data.playerWins,
          correctCount: data.correctCount,
          totalModels: data.totalModels,
          duration: data.duration,
          score: data.score
        });
        setGameMessage(data.message || '');
        setGameState('finished');
        fetchStats();
        fetchLeaderboard();
      } else if (data.type === 'gameResult') {
        if (data.modelStates) {
          setModelHistory(data.modelStates);
        }

        if (!data.gameOver) {
          setCurrentRound(data.nextRound);
          setTimeout(() => {
            const outputs = {};
            config.models.forEach(model => {
              if (data.modelStates && !data.modelStates[model.name]?.correct) {
                outputs[model.name] = '';
              } else {
                outputs[model.name] = modelOutputs[model.name];
              }
            });
            setModelOutputs(outputs);
          }, 2000);
        }
      }
    };

    ws.current.onerror = (error) => {
      console.error('WebSocket error:', error);
    };
  };

  const startGame = () => {
    if (!riddle || !answer || clues.some(c => !c)) {
      alert('Please fill in all fields');
      return;
    }

    connectWebSocket();
    
    ws.current.onopen = () => {
      const submission = {
        riddle,
        answer,
        clues: clues.filter(c => c.trim()),
        difficulty
      };
      
      ws.current.send(JSON.stringify(submission));
      setGameState('playing');
      
      const outputs = {};
      config.models.forEach(model => {
        outputs[model.name] = '';
      });
      setModelOutputs(outputs);
      setModelResults({});
      setModelHistory({});
      setCurrentRound(0);
    };
  };

  const resetGame = () => {
    setGameState('setup');
    setRiddle('');
    setAnswer('');
    setClues(['', '', '']);
    setDifficulty('medium');
    
    const outputs = {};
    if (config) {
      config.models.forEach(model => {
        outputs[model.name] = '';
      });
    }
    setModelOutputs(outputs);
    setModelResults({});
    setModelHistory({});
    setCurrentRound(0);
    setGameResult(null);
    setGameMessage('');
    
    if (ws.current) {
      ws.current.close();
    }
  };

  const getModelColor = (index) => {
    const colors = [
      'from-green-500 to-emerald-600',
      'from-orange-500 to-amber-600',
      'from-blue-500 to-indigo-600',
      'from-purple-500 to-pink-600',
      'from-red-500 to-rose-600',
      'from-cyan-500 to-teal-600'
    ];
    return colors[index % colors.length];
  };

  const getModelIcon = (provider) => {
    const icons = {
      'openai': '🤖',
      'anthropic': '🧠',
      'google': '✨',
      'ollama': '🦙',
      'huggingface': '🤗'
    };
    return icons[provider] || '🎯';
  };

  const getDifficultyColor = (diff) => {
    const colors = {
      'easy': 'bg-green-600',
      'medium': 'bg-yellow-600',
      'hard': 'bg-red-600'
    };
    return colors[diff] || 'bg-gray-600';
  };

  const ModelColumn = ({ model, output, isCorrect, index }) => {
    const history = modelHistory[model.name];
    const hasWon = history?.correct;
    const isThinking = output && output.length > 0 && isCorrect === undefined;
    
    return (
      <div className="flex-1 bg-gray-800 rounded-lg p-6 flex flex-col min-w-0">
        <div className={`bg-gradient-to-r ${getModelColor(index)} rounded-lg p-4 mb-4`}>
          <div className="text-4xl mb-2 text-center">{getModelIcon(model.provider)}</div>
          <h3 className="text-xl font-bold text-white text-center uppercase tracking-wide">
            {model.name}
          </h3>
          <p className="text-xs text-white text-center opacity-75 mt-1">
            {model.model}
          </p>
          {hasWon && (
            <div className="mt-2 bg-green-500 text-white text-center py-1 rounded text-xs font-bold">
              WON IN ROUND {history.round}
            </div>
          )}
        </div>
        
        {history?.allGuesses && history.allGuesses.length > 0 && (
          <div className="mb-4 space-y-2">
            {history.allGuesses.map((guess, idx) => {
              const isCurrentGuess = idx === history.allGuesses.length - 1 && !hasWon;
              if (isCurrentGuess && isThinking) return null;
              
              return (
                <div key={idx} className="bg-gray-900 rounded-lg p-3 flex items-start">
                  <div className={`mr-2 mt-1 flex-shrink-0 ${history.guessResults[idx] ? 'text-green-400' : 'text-red-400'}`}>
                    {history.guessResults[idx] ? '✓' : '✗'}
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-center mb-1">
                      <div className="text-gray-500 text-xs">Round {idx + 1}</div>
                      {history.responseTimes && history.responseTimes[idx] && (
                        <div className="text-gray-500 text-xs">
                          {history.responseTimes[idx].toFixed(2)}s
                        </div>
                      )}
                    </div>
                    <div className="text-gray-300 text-sm break-words">{guess}</div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
        
        {!hasWon && (
          <div className="flex-1 bg-gray-900 rounded-lg p-4 mb-4 overflow-auto min-h-[200px] relative">
            <div className="text-gray-300 whitespace-pre-wrap font-mono text-sm break-words">
              {output ? output : <span className="text-gray-600">Waiting for response...</span>}
            </div>
            {isThinking && (
              <div className="absolute bottom-2 right-2">
                <div className="flex items-center gap-2 text-green-400 text-xs">
                  <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                  AI thinking...
                </div>
              </div>
            )}
          </div>
        )}
        
        {isCorrect !== undefined && !hasWon && (
          <div className={`rounded-lg p-3 flex items-center justify-center ${
            isCorrect ? 'bg-green-600' : 'bg-red-600'
          }`}>
            {isCorrect ? (
              <><Trophy className="mr-2" size={20} /> CORRECT</>
            ) : (
              <><XCircle className="mr-2" size={20} /> INCORRECT</>
            )}
          </div>
        )}
      </div>
    );
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 flex items-center justify-center">
        <div className="text-white text-xl">Loading configuration...</div>
      </div>
    );
  }

  if (!config || !config.models || config.models.length === 0) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 flex items-center justify-center p-8">
        <div className="bg-red-900 border border-red-700 rounded-lg p-8 max-w-2xl">
          <AlertCircle className="text-red-400 mb-4" size={48} />
          <h2 className="text-white text-2xl font-bold mb-4">Configuration Error</h2>
          <p className="text-gray-300 mb-4">
            No models configured. Please create a config.json file with your model settings.
          </p>
          <p className="text-gray-400 text-sm">
            The server should have a config.json file in the same directory.
          </p>
        </div>
      </div>
    );
  }

  if (gameState === 'setup') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8">
        <div className="max-w-3xl mx-auto">
          <div className="mb-6 flex justify-between items-center">
            <div className="flex gap-2">
              <button
                onClick={() => setGameState('stats')}
                className="bg-gray-700 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition flex items-center gap-2"
              >
                <BarChart3 size={18} />
                Stats
              </button>
              <button
                onClick={() => setGameState('leaderboard')}
                className="bg-gray-700 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition flex items-center gap-2"
              >
                <Award size={18} />
                Leaderboard
              </button>
            </div>
          </div>

          <div className="text-center mb-8">
            <h1 className="text-5xl font-bold text-white mb-4">Create Your Riddle</h1>
            <p className="text-gray-300 text-lg">
              Stump some AIs but not all of them to win!
            </p>
          </div>

          <div className="bg-gray-800 rounded-lg p-8 shadow-2xl">
            <div className="mb-6">
              <label className="block text-white font-semibold mb-2">
                Difficulty Level
              </label>
              <div className="flex gap-3">
                {['easy', 'medium', 'hard'].map(diff => (
                  <button
                    key={diff}
                    onClick={() => setDifficulty(diff)}
                    className={`flex-1 py-3 rounded-lg font-bold uppercase transition ${
                      difficulty === diff
                        ? getDifficultyColor(diff) + ' text-white'
                        : 'bg-gray-700 text-gray-400 hover:bg-gray-600'
                    }`}
                  >
                    {diff}
                  </button>
                ))}
              </div>
            </div>

            <div className="mb-6">
              <label className="block text-white font-semibold mb-2">
                Your Riddle
              </label>
              <textarea
                value={riddle}
                onChange={(e) => setRiddle(e.target.value)}
                className="w-full bg-gray-700 text-white rounded-lg p-3 min-h-[100px] focus:ring-2 focus:ring-purple-500 outline-none"
                placeholder="Enter your riddle here..."
              />
            </div>

            <div className="mb-6">
              <label className="block text-white font-semibold mb-2">
                Answer
              </label>
              <input
                type="text"
                value={answer}
                onChange={(e) => setAnswer(e.target.value)}
                className="w-full bg-gray-700 text-white rounded-lg p-3 focus:ring-2 focus:ring-purple-500 outline-none"
                placeholder="The correct answer..."
              />
            </div>

            <div className="mb-6">
              <label className="block text-white font-semibold mb-2">
                Clues (will be revealed one per round)
              </label>
              {clues.map((clue, i) => (
                <input
                  key={i}
                  type="text"
                  value={clue}
                  onChange={(e) => {
                    const newClues = [...clues];
                    newClues[i] = e.target.value;
                    setClues(newClues);
                  }}
                  className="w-full bg-gray-700 text-white rounded-lg p-3 mb-3 focus:ring-2 focus:ring-purple-500 outline-none"
                  placeholder={`Clue ${i + 1}...`}
                />
              ))}
            </div>

            <button
              onClick={startGame}
              className="w-full bg-gradient-to-r from-purple-600 to-pink-600 text-white font-bold py-4 rounded-lg hover:from-purple-700 hover:to-pink-700 transition flex items-center justify-center text-lg"
            >
              <Zap className="mr-2" size={24} />
              Start Game
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (gameState === 'stats') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8">
        <div className="max-w-4xl mx-auto">
          <div className="mb-6">
            <button
              onClick={() => setGameState('setup')}
              className="bg-gray-700 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition"
            >
              Back
            </button>
          </div>

          <div className="text-center mb-8">
            <h1 className="text-5xl font-bold text-white mb-2 flex items-center justify-center">
              <BarChart3 className="mr-3" size={48} />
              Your Statistics
            </h1>
          </div>

          {stats ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-gray-800 rounded-lg p-6">
                <h3 className="text-white font-bold text-xl mb-4">Overall Stats</h3>
                <div className="space-y-3">
                  <div className="flex justify-between">
                    <span className="text-gray-400">Total Games:</span>
                    <span className="text-white font-bold">{stats.totalGames || 0}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Wins:</span>
                    <span className="text-green-400 font-bold">{stats.wins || 0}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Losses:</span>
                    <span className="text-red-400 font-bold">{stats.losses || 0}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Win Rate:</span>
                    <span className="text-purple-400 font-bold">{(stats.winRate || 0).toFixed(1)}%</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-400">Avg Duration:</span>
                    <span className="text-white font-bold">{(stats.averageDuration || 0).toFixed(1)}s</span>
                  </div>
                </div>
              </div>

              <div className="bg-gray-800 rounded-lg p-6">
                <h3 className="text-white font-bold text-xl mb-4">By Difficulty</h3>
                <div className="space-y-3">
                  {stats.byDifficulty && Object.keys(stats.byDifficulty).length > 0 ? (
                    Object.entries(stats.byDifficulty).map(([diff, count]) => (
                      <div key={diff} className="flex justify-between items-center">
                        <span className="text-gray-400 capitalize">{diff}:</span>
                        <span className={`${getDifficultyColor(diff)} text-white px-3 py-1 rounded text-sm font-bold`}>
                          {count} games
                        </span>
                      </div>
                    ))
                  ) : (
                    <p className="text-gray-400 text-center py-4">No games played yet</p>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-gray-800 rounded-lg p-12 text-center">
              <p className="text-gray-400 text-lg">Loading statistics...</p>
            </div>
          )}
        </div>
      </div>
    );
  }

  if (gameState === 'leaderboard') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8">
        <div className="max-w-6xl mx-auto">
          <div className="mb-6">
            <button
              onClick={() => setGameState('setup')}
              className="bg-gray-700 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition"
            >
              Back
            </button>
          </div>

          <div className="text-center mb-8">
            <h1 className="text-5xl font-bold text-white mb-2 flex items-center justify-center">
              <Award className="mr-3" size={48} />
              Leaderboard
            </h1>
            <p className="text-gray-400">Top 100 scoring riddles</p>
          </div>

          <div className="bg-gray-800 rounded-lg overflow-hidden">
            {leaderboard && leaderboard.length > 0 ? (
              <table className="w-full">
                <thead className="bg-gray-900">
                  <tr>
                    <th className="text-left p-4 text-gray-400 font-semibold">Rank</th>
                    <th className="text-left p-4 text-gray-400 font-semibold">Score</th>
                    <th className="text-left p-4 text-gray-400 font-semibold">Difficulty</th>
                    <th className="text-left p-4 text-gray-400 font-semibold">Result</th>
                    <th className="text-left p-4 text-gray-400 font-semibold">Riddle</th>
                    <th className="text-left p-4 text-gray-400 font-semibold">Duration</th>
                  </tr>
                </thead>
                <tbody>
                  {leaderboard.map((entry, index) => (
                    <tr key={index} className="border-t border-gray-700 hover:bg-gray-750">
                      <td className="p-4">
                        <span className={`font-bold ${index < 3 ? 'text-yellow-400 text-xl' : 'text-white'}`}>
                          {index === 0 ? 'Gold' : index === 1 ? 'Silver' : index === 2 ? 'Bronze' : `#${index + 1}`}
                        </span>
                      </td>
                      <td className="p-4 text-white font-bold text-lg">{entry.score}</td>
                      <td className="p-4">
                        <span className={`${getDifficultyColor(entry.difficulty)} text-white px-2 py-1 rounded text-xs uppercase`}>
                          {entry.difficulty}
                        </span>
                      </td>
                      <td className="p-4">
                        <span className={entry.playerWon ? 'text-green-400' : 'text-red-400'}>
                          {entry.correctCount}/{entry.totalModels}
                        </span>
                      </td>
                      <td className="p-4 text-gray-300 max-w-md truncate">{entry.riddle}</td>
                      <td className="p-4 text-gray-400">{entry.duration.toFixed(1)}s</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <div className="text-center py-12 text-gray-400">
                No games played yet. Be the first to set a score!
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }

  if (gameState === 'playing') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-6">
            <h1 className="text-4xl font-bold text-white mb-2">Turing Roulette</h1>
            <div className="flex justify-center gap-4">
              <div className="bg-gray-800 rounded-lg px-4 py-2">
                <p className="text-purple-400 font-semibold">Round {currentRound + 1}</p>
              </div>
              <div className={`${getDifficultyColor(difficulty)} rounded-lg px-4 py-2`}>
                <p className="text-white font-semibold uppercase">{difficulty}</p>
              </div>
            </div>
          </div>

          <div className="bg-gray-800 rounded-lg p-6 mb-6">
            <h2 className="text-white font-bold text-xl mb-2">Your Riddle:</h2>
            <p className="text-gray-300 text-lg italic">{riddle}</p>
            {currentRound > 0 && (
              <div className="mt-4">
                <h3 className="text-purple-400 font-semibold mb-2">Clues Given:</h3>
                {clues.slice(0, currentRound).map((clue, i) => (
                  <p key={i} className="text-gray-300 ml-4">• {clue}</p>
                ))}
              </div>
            )}
          </div>

          <div className="flex gap-6 overflow-x-auto pb-4">
            {config.models.map((model, i) => (
              <ModelColumn 
                key={i}
                model={model}
                output={modelOutputs[model.name]}
                isCorrect={modelResults[model.name]}
                index={i}
              />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (gameState === 'finished') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8 flex items-center justify-center">
        <div className="max-w-2xl bg-gray-800 rounded-lg p-12 text-center shadow-2xl">
          <div className="text-6xl mb-6">
            {gameResult.playerWins ? '🎉' : '😔'}
          </div>
          <h1 className={`text-5xl font-bold mb-4 ${
            gameResult.playerWins ? 'text-green-400' : 'text-red-400'
          }`}>
            {gameResult.playerWins ? 'YOU WIN!' : 'YOU LOSE!'}
          </h1>
          
          <div className="bg-gray-900 rounded-lg p-6 mb-6">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-gray-400 text-sm">Score</p>
                <p className="text-white font-bold text-3xl">{gameResult.score}</p>
              </div>
              <div>
                <p className="text-gray-400 text-sm">Duration</p>
                <p className="text-white font-bold text-3xl">{gameResult.duration.toFixed(1)}s</p>
              </div>
              <div>
                <p className="text-gray-400 text-sm">Correct Models</p>
                <p className="text-white font-bold text-3xl">{gameResult.correctCount}/{gameResult.totalModels}</p>
              </div>
              <div>
                <p className="text-gray-400 text-sm">Difficulty</p>
                <p className={`${getDifficultyColor(difficulty)} inline-block px-3 py-1 rounded text-white font-bold uppercase mt-1`}>
                  {difficulty}
                </p>
              </div>
            </div>
          </div>

          <p className="text-gray-400 mb-8">
            {gameMessage || (
              gameResult.playerWins
                ? 'Perfect! You stumped some AIs but not all of them!'
                : gameResult.correctCount === gameResult.totalModels
                  ? 'All models guessed correctly. Your riddle was too easy!'
                  : 'None of the models got it right even with all the clues!'
            )}
          </p>
          
          <div className="flex gap-4">
            <button
              onClick={resetGame}
              className="flex-1 bg-gradient-to-r from-purple-600 to-pink-600 text-white font-bold py-4 px-8 rounded-lg hover:from-purple-700 hover:to-pink-700 transition text-lg"
            >
              Play Again
            </button>
            <button
              onClick={() => setGameState('setup')}
              className="flex-1 bg-gray-700 text-white font-bold py-4 px-8 rounded-lg hover:bg-gray-600 transition text-lg"
            >
              New Riddle
            </button>
          </div>
        </div>
      </div>
    );
  }
}