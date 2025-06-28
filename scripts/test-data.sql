-- Test data for fern-platform database
-- This script creates sample projects, test runs, suite runs, and spec runs

-- Clear existing test data (optional - comment out if you want to keep existing data)
-- DELETE FROM spec_runs WHERE test_run_id IN (SELECT id FROM test_runs WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003'));
-- DELETE FROM suite_runs WHERE test_run_id IN (SELECT id FROM test_runs WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003'));
-- DELETE FROM test_runs WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003');
-- DELETE FROM project_details WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003');

-- Insert sample projects
INSERT INTO project_details (project_id, name, description, repository, default_branch, team, is_active, created_at, updated_at) VALUES
('project-frontend-001', 'E-Commerce Frontend', 'React-based e-commerce platform frontend with TypeScript', 'https://github.com/example/ecommerce-frontend', 'main', 'frontend', true, NOW() - INTERVAL '30 days', NOW()),
('project-backend-002', 'API Gateway Service', 'High-performance API gateway built with Go', 'https://github.com/example/api-gateway', 'main', 'backend', true, NOW() - INTERVAL '45 days', NOW()),
('project-mobile-003', 'Mobile Banking App', 'Cross-platform mobile banking application using React Native', 'https://github.com/example/mobile-banking', 'develop', 'mobile', true, NOW() - INTERVAL '60 days', NOW());

-- Function to generate test data
DO $$
DECLARE
    project_rec RECORD;
    test_run_id INTEGER;
    suite_run_id INTEGER;
    suite_counter INTEGER;
    spec_counter INTEGER;
    total_tests INTEGER;
    passed_tests INTEGER;
    failed_tests INTEGER;
    skipped_tests INTEGER;
    test_status VARCHAR;
    run_duration INTEGER;
    spec_duration INTEGER;
BEGIN
    -- Loop through each project
    FOR project_rec IN 
        SELECT project_id, name 
        FROM project_details 
        WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
    LOOP
        -- Create 3-5 test runs per project
        FOR i IN 1..3 + floor(random() * 3)::int LOOP
            total_tests := 0;
            passed_tests := 0;
            failed_tests := 0;
            skipped_tests := 0;
            run_duration := 0;
            
            -- Insert test run
            INSERT INTO test_runs (
                project_id, run_id, status, branch, commit_sha, 
                start_time, end_time, duration_ms, total_tests, 
                passed_tests, failed_tests, skipped_tests, environment,
                created_at, updated_at
            ) VALUES (
                project_rec.project_id,
                gen_random_uuid()::text,
                'completed',
                CASE 
                    WHEN random() < 0.7 THEN 'main'
                    WHEN random() < 0.9 THEN 'develop'
                    ELSE 'feature/test-' || floor(random() * 100)::text
                END,
                substring(md5(random()::text) from 1 for 8),
                NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int),
                NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int) + INTERVAL '1 minute' * (5 + floor(random() * 20)::int),
                0, -- Will update later
                0, -- Will update later
                0, -- Will update later
                0, -- Will update later
                0, -- Will update later
                CASE 
                    WHEN random() < 0.5 THEN 'production'
                    WHEN random() < 0.8 THEN 'staging'
                    ELSE 'development'
                END,
                NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int),
                NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int)
            ) RETURNING id INTO test_run_id;
            
            -- Create 2-5 test suites per test run
            suite_counter := 2 + floor(random() * 4)::int;
            FOR j IN 1..suite_counter LOOP
                -- Insert suite run
                INSERT INTO suite_runs (
                    test_run_id, suite_name, status, start_time, end_time,
                    duration_ms, total_tests, passed_tests, failed_tests,
                    skipped_tests, created_at, updated_at
                ) VALUES (
                    test_run_id,
                    CASE project_rec.project_id
                        WHEN 'project-frontend-001' THEN 
                            CASE j
                                WHEN 1 THEN 'Component Tests'
                                WHEN 2 THEN 'Integration Tests'
                                WHEN 3 THEN 'E2E Tests'
                                WHEN 4 THEN 'Unit Tests'
                                ELSE 'Smoke Tests'
                            END
                        WHEN 'project-backend-002' THEN
                            CASE j
                                WHEN 1 THEN 'API Tests'
                                WHEN 2 THEN 'Performance Tests'
                                WHEN 3 THEN 'Security Tests'
                                WHEN 4 THEN 'Database Tests'
                                ELSE 'Load Tests'
                            END
                        ELSE -- project-mobile-003
                            CASE j
                                WHEN 1 THEN 'UI Tests'
                                WHEN 2 THEN 'Device Tests'
                                WHEN 3 THEN 'Network Tests'
                                WHEN 4 THEN 'Offline Tests'
                                ELSE 'Platform Tests'
                            END
                    END,
                    'completed',
                    NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int),
                    NOW() - INTERVAL '1 day' * (i + floor(random() * 10)::int) + INTERVAL '1 second' * (30 + floor(random() * 300)::int),
                    0, -- Will update later
                    0, -- Will update later
                    0, -- Will update later
                    0, -- Will update later
                    0, -- Will update later
                    NOW(),
                    NOW()
                ) RETURNING id INTO suite_run_id;
                
                -- Create 5-10 test specs per suite
                spec_counter := 5 + floor(random() * 6)::int;
                FOR k IN 1..spec_counter LOOP
                    -- Determine test status based on project
                    -- project-backend-002 should have ~50% pass rate
                    IF project_rec.project_id = 'project-backend-002' THEN
                        IF random() < 0.5 THEN
                            test_status := 'passed';
                            passed_tests := passed_tests + 1;
                        ELSIF random() < 0.8 THEN
                            test_status := 'failed';
                            failed_tests := failed_tests + 1;
                        ELSE
                            test_status := 'skipped';
                            skipped_tests := skipped_tests + 1;
                        END IF;
                    ELSE
                        -- Other projects have better pass rates
                        IF random() < 0.8 THEN
                            test_status := 'passed';
                            passed_tests := passed_tests + 1;
                        ELSIF random() < 0.95 THEN
                            test_status := 'failed';
                            failed_tests := failed_tests + 1;
                        ELSE
                            test_status := 'skipped';
                            skipped_tests := skipped_tests + 1;
                        END IF;
                    END IF;
                    
                    total_tests := total_tests + 1;
                    spec_duration := 100 + floor(random() * 5000)::int;
                    run_duration := run_duration + spec_duration;
                    
                    -- Insert spec run
                    INSERT INTO spec_runs (
                        suite_run_id, spec_name, file_path, line_number,
                        status, duration_ms, error_message, stack_trace,
                        created_at, updated_at
                    ) VALUES (
                        suite_run_id,
                        CASE project_rec.project_id
                            WHEN 'project-frontend-001' THEN 
                                'should ' || 
                                CASE floor(random() * 10)::int
                                    WHEN 0 THEN 'render component correctly'
                                    WHEN 1 THEN 'handle user interactions'
                                    WHEN 2 THEN 'validate form inputs'
                                    WHEN 3 THEN 'display error messages'
                                    WHEN 4 THEN 'update state on action'
                                    WHEN 5 THEN 'fetch data from API'
                                    WHEN 6 THEN 'handle loading states'
                                    WHEN 7 THEN 'navigate between pages'
                                    WHEN 8 THEN 'apply correct styling'
                                    ELSE 'handle edge cases'
                                END
                            WHEN 'project-backend-002' THEN
                                'should ' ||
                                CASE floor(random() * 10)::int
                                    WHEN 0 THEN 'route requests correctly'
                                    WHEN 1 THEN 'authenticate users'
                                    WHEN 2 THEN 'validate request payload'
                                    WHEN 3 THEN 'handle rate limiting'
                                    WHEN 4 THEN 'return correct status codes'
                                    WHEN 5 THEN 'transform response data'
                                    WHEN 6 THEN 'handle timeouts gracefully'
                                    WHEN 7 THEN 'cache responses appropriately'
                                    WHEN 8 THEN 'log requests properly'
                                    ELSE 'handle concurrent requests'
                                END
                            ELSE -- project-mobile-003
                                'should ' ||
                                CASE floor(random() * 10)::int
                                    WHEN 0 THEN 'render on all screen sizes'
                                    WHEN 1 THEN 'handle touch gestures'
                                    WHEN 2 THEN 'work offline'
                                    WHEN 3 THEN 'sync data when online'
                                    WHEN 4 THEN 'handle biometric auth'
                                    WHEN 5 THEN 'receive push notifications'
                                    WHEN 6 THEN 'access device camera'
                                    WHEN 7 THEN 'handle deep links'
                                    WHEN 8 THEN 'persist user preferences'
                                    ELSE 'handle app lifecycle'
                                END
                        END || ' - test #' || k,
                        'src/tests/suite_' || j || '/test_' || k || '.spec.ts',
                        10 + floor(random() * 200)::int,
                        test_status,
                        spec_duration,
                        CASE 
                            WHEN test_status = 'failed' THEN 'Expected value to be true but got false'
                            ELSE NULL
                        END,
                        CASE 
                            WHEN test_status = 'failed' THEN 'Error: Assertion failed\n  at Object.<anonymous> (test.spec.ts:' || (10 + floor(random() * 200)::int)::text || ':15)\n  at runTest (runner.js:123:5)'
                            ELSE NULL
                        END,
                        NOW(),
                        NOW()
                    );
                END LOOP;
                
                -- Update suite run with aggregated data
                UPDATE suite_runs 
                SET 
                    duration_ms = (run_duration / suite_counter),
                    total_tests = (total_tests / suite_counter),
                    passed_tests = (passed_tests / suite_counter),
                    failed_tests = (failed_tests / suite_counter),
                    skipped_tests = (skipped_tests / suite_counter)
                WHERE id = suite_run_id;
            END LOOP;
            
            -- Update test run with aggregated data
            UPDATE test_runs 
            SET 
                duration_ms = run_duration,
                total_tests = total_tests,
                passed_tests = passed_tests,
                failed_tests = failed_tests,
                skipped_tests = skipped_tests,
                status = CASE 
                    WHEN failed_tests > 0 THEN 'failed'
                    WHEN skipped_tests = total_tests THEN 'skipped'
                    ELSE 'passed'
                END
            WHERE id = test_run_id;
        END LOOP;
    END LOOP;
END $$;

-- Add some recent test runs for better dashboard display
DO $$
DECLARE
    project_rec RECORD;
    test_run_id INTEGER;
    suite_run_id INTEGER;
BEGIN
    -- Add very recent test runs (within last 24 hours)
    FOR project_rec IN 
        SELECT project_id, name 
        FROM project_details 
        WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
    LOOP
        -- Insert a recent test run
        INSERT INTO test_runs (
            project_id, run_id, status, branch, commit_sha, 
            start_time, end_time, duration_ms, total_tests, 
            passed_tests, failed_tests, skipped_tests, environment,
            created_at, updated_at
        ) VALUES (
            project_rec.project_id,
            gen_random_uuid()::text,
            'running',
            'main',
            substring(md5(random()::text) from 1 for 8),
            NOW() - INTERVAL '30 minutes',
            NULL,
            0,
            0,
            0,
            0,
            0,
            'production',
            NOW() - INTERVAL '30 minutes',
            NOW()
        ) RETURNING id INTO test_run_id;
        
        -- Add one suite that's running
        INSERT INTO suite_runs (
            test_run_id, suite_name, status, start_time, end_time,
            duration_ms, total_tests, passed_tests, failed_tests,
            skipped_tests, created_at, updated_at
        ) VALUES (
            test_run_id,
            'Integration Tests',
            'running',
            NOW() - INTERVAL '30 minutes',
            NULL,
            0,
            0,
            0,
            0,
            0,
            NOW(),
            NOW()
        );
    END LOOP;
END $$;

-- Display summary
SELECT 
    'Projects created' as metric,
    COUNT(DISTINCT project_id) as count
FROM project_details
WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
UNION ALL
SELECT 
    'Test runs created',
    COUNT(*)
FROM test_runs
WHERE project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
UNION ALL
SELECT 
    'Suite runs created',
    COUNT(*)
FROM suite_runs sr
JOIN test_runs tr ON sr.test_run_id = tr.id
WHERE tr.project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
UNION ALL
SELECT 
    'Spec runs created',
    COUNT(*)
FROM spec_runs spr
JOIN suite_runs sr ON spr.suite_run_id = sr.id
JOIN test_runs tr ON sr.test_run_id = tr.id
WHERE tr.project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003');

-- Show pass rates by project
SELECT 
    pd.name as project_name,
    COUNT(DISTINCT tr.id) as total_test_runs,
    SUM(tr.total_tests) as total_tests,
    SUM(tr.passed_tests) as passed_tests,
    SUM(tr.failed_tests) as failed_tests,
    SUM(tr.skipped_tests) as skipped_tests,
    ROUND(CASE 
        WHEN SUM(tr.total_tests) > 0 
        THEN (SUM(tr.passed_tests)::numeric / SUM(tr.total_tests)::numeric) * 100
        ELSE 0 
    END, 2) as pass_rate_percentage
FROM project_details pd
JOIN test_runs tr ON pd.project_id = tr.project_id
WHERE pd.project_id IN ('project-frontend-001', 'project-backend-002', 'project-mobile-003')
GROUP BY pd.name
ORDER BY pd.name;