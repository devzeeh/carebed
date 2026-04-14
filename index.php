<?php

// 1. Load the Composer autoloader and .env file
// Parse the file into an associative array
$config = parse_ini_file(__DIR__ . '/config.ini');

$apiKey = $config['GEMINI_API_KEY'] ?? null;
if (!$apiKey) {
    die("Error: API key not found in config.ini\n");
}

// 2. Define the mock sensor data
$sensorData = [
    "patient_age" => 45,
    "current_bpm" => 115,
    "state" => "resting",
    "duration_minutes" => 15,
    "symptoms" => ["mild shortness of breath", "slight dizziness"]
];

$userPrompt = "Analyze the following sensor data:\n" . json_encode($sensorData, JSON_PRETTY_PRINT);

// 3. Construct the API Payload
$payload = [
    // Persona and Guardrails
    "systemInstruction" => [
        "parts" => [
            ["text" => "You are Carebed AI, a backend analysis engine for a health monitoring app. " .
                       "Analyze incoming heart rate sensor data alongside patient context. " .
                       "Evaluate the BPM, suggest potential benign causes, and list potential illnesses. " .
                       "CRITICAL RULE: You are not a doctor. You must NEVER make a definitive diagnosis. " .
                       "Always include a disclaimer advising the user to seek professional medical help."]
        ]
    ],
    // The User's Data
    "contents" => [
        [
            "parts" => [
                ["text" => $userPrompt]
            ]
        ]
    ],
    // Configuration & Strict JSON Schema Enforcement
    "generationConfig" => [
        "temperature" => 0.1, // Highly analytical, low hallucination
        "responseMimeType" => "application/json",
        "responseSchema" => [
            "type" => "OBJECT",
            "properties" => [
                "analysis" => ["type" => "STRING", "description" => "Brief summary of what the HR indicates."],
                "severity_level" => ["type" => "STRING", "description" => "Severity level: Normal, Low, Moderate, or High."],
                "potential_benign_causes" => [
                    "type" => "ARRAY", 
                    "items" => ["type" => "STRING"]
                ],
                "potential_illnesses" => [
                    "type" => "ARRAY", 
                    "items" => ["type" => "STRING"]
                ],
                "action_plan" => ["type" => "STRING", "description" => "Immediate, safe steps the user should take."],
                "medical_disclaimer" => ["type" => "STRING", "description" => "Mandatory disclaimer stating this is not medical advice."]
            ],
            "required" => ["analysis", "severity_level", "potential_benign_causes", "potential_illnesses", "action_plan", "medical_disclaimer"]
        ]
    ]
];

// 4. Set up the cURL Request to the Gemini REST API
$url = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" . $apiKey;

echo "Sending data to Carebed AI for analysis...\n\n";

$ch = curl_init($url);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json'
]);
// Convert our PHP payload array to a JSON string
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));

// 5. Execute the request
$response = curl_exec($ch);

if (curl_errno($ch)) {
    die('cURL error: ' . curl_error($ch));
}
curl_close($ch);

// 6. Parse and display the response
$responseData = json_decode($response, true);

// Navigate through the REST response structure to get the actual AI text
if (isset($responseData['candidates'][0]['content']['parts'][0]['text'])) {
    $aiJsonString = $responseData['candidates'][0]['content']['parts'][0]['text'];
    
    // Decode and re-encode to pretty-print the JSON output
    $formattedOutput = json_encode(json_decode($aiJsonString), JSON_PRETTY_PRINT);
    echo $formattedOutput . "\n";
} else {
    echo "Unexpected API Response:\n";
    print_r($responseData);
}