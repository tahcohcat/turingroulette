import React, { useState, useEffect, useRef } from 'react';
import { Brain, Zap, AlertCircle, Trophy, XCircle, Settings } from 'lucide-react';

export default function TuringRoulette() {
  const [gameState, setGameState] = useState('setup');
  const [riddle, setRiddle] = useState('');
  const [answer, setAnswer] = useState('');
  const [clues, setClues] = useState(['', '', '']);
  const [modelOutputs, setModelOutputs] = useState({});
  const [modelResults, setModelResults] = useState({});
  const [currentRound, setCurrentRound] = useState(0);
  const [gameResult, setGameResult] = useState(null);
  const [config, setConfig] = useState(null);
  const [loading, setLoading] = useState(true);
  const ws = useRef(null);

  useEffect(() => {
    fetchConfig();
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
      } else if (data.type === 'gameResult') {
        if (data.gameOver) {
          setGameResult({
            playerWins: data.playerWins,
            correctCount: data.correctCount,
            totalModels: data.totalModels
          });
          setGameState('finished');
        } else {
          setCurrentRound(data.nextRound);
          setTimeout(() => {
            const outputs = {};
            config.models.forEach(model => {
              outputs[model.name] = '';
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
        clues: clues.filter(c => c.trim())
      };
      
      ws.current.send(JSON.stringify(submission));
      setGameState('playing');
      
      const outputs = {};
      config.models.forEach(model => {
        outputs[model.name] = '';
      });
      setModelOutputs(outputs);
      setModelResults({});
      setCurrentRound(0);
    };
  };

  const resetGame = () => {
    setGameState('setup');
    setRiddle('');
    setAnswer('');
    setClues(['', '', '']);
    
    const outputs = {};
    config.models.forEach(model => {
      outputs[model.name] = '';
    });
    setModelOutputs(outputs);
    setModelResults({});
    setCurrentRound(0);
    setGameResult(null);
    
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
      'ollama': '🦙'
    };
    return icons[provider] || '🎯';
  };

  const ModelColumn = ({ model, output, isCorrect, index }) => {
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
        </div>
        
        <div className="flex-1 bg-gray-900 rounded-lg p-4 mb-4 overflow-auto min-h-[200px] relative">
          <div className="text-gray-300 whitespace-pre-wrap font-mono text-sm break-words">
            {output || <span className="text-gray-600">Waiting for response...</span>}
          </div>
          {output && (
            <div className="absolute bottom-2 right-2">
              <div className="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
            </div>
          )}
        </div>
        
        {isCorrect !== undefined && (
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
          <div className="text-center mb-8">
            <h1 className="text-5xl font-bold text-white mb-4 flex items-center justify-center">
              <Brain className="mr-4" size={48} />
              Turing Roulette
            </h1>
            <p className="text-gray-300 text-lg">
              Stump some AIs but not all of them to win!
            </p>
          </div>

          <div className="bg-gray-800 rounded-lg p-6 mb-6">
            <div className="flex items-center mb-4">
              <Settings className="mr-2 text-purple-400" size={20} />
              <h3 className="text-white font-semibold">Active Models</h3>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
              {config.models.map((model, i) => (
                <div key={i} className="bg-gray-900 rounded-lg p-3 text-center">
                  <div className="text-2xl mb-1">{getModelIcon(model.provider)}</div>
                  <div className="text-white font-semibold text-sm">{model.name}</div>
                  <div className="text-gray-400 text-xs">{model.provider}</div>
                </div>
              ))}
            </div>
          </div>

          <div className="bg-gray-800 rounded-lg p-8 shadow-2xl">
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

  if (gameState === 'playing') {
    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900 p-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-6">
            <h1 className="text-4xl font-bold text-white mb-2">Turing Roulette</h1>
            <div className="bg-gray-800 rounded-lg p-4 inline-block">
              <p className="text-purple-400 font-semibold">Round {currentRound + 1}</p>
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
          <div className={`text-6xl mb-6`}>
            {gameResult.playerWins ? '🎉' : '😔'}
          </div>
          <h1 className={`text-5xl font-bold mb-4 ${
            gameResult.playerWins ? 'text-green-400' : 'text-red-400'
          }`}>
            {gameResult.playerWins ? 'YOU WIN!' : 'GAME OVER'}
          </h1>
          <p className="text-gray-300 text-xl mb-8">
            {gameResult.correctCount} out of {gameResult.totalModels} models got it right
          </p>
          <p className="text-gray-400 mb-8">
            {gameResult.playerWins 
              ? 'You stumped some AIs but not all of them! Perfect balance!' 
              : gameResult.correctCount === gameResult.totalModels 
                ? 'All models guessed correctly. Your riddle was too easy!'
                : 'None of the models got it right even with all the clues!'}
          </p>
          <button
            onClick={resetGame}
            className="bg-gradient-to-r from-purple-600 to-pink-600 text-white font-bold py-4 px-8 rounded-lg hover:from-purple-700 hover:to-pink-700 transition text-lg"
          >
            Play Again
          </button>
        </div>
      </div>
    );
  }
}